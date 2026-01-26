// ===========================================================================================
// Provides functions to interact with Azure RBAC PIM API
// for managing Privileged Identity Management (PIM) group activations.
//
// Why not use Microsoft Graph API? Because it requires permissions not available via
// Azure CLI authentication (e.g. PrivilegedAccess.ReadWrite.AzureADGroup)
// And various other reasons that make it impractical for any real world usage.
// ===========================================================================================

package pim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const (
	pimAPIScope   = "https://api.azrbac.mspim.azure.com/.default"
	pimAPIBaseURL = "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess/aadGroups"
)

// PIM API response structures
type pimResource struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
}

type pimRoleDefinition struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type pimRoleAssignment struct {
	ID                string            `json:"id"`
	ResourceID        string            `json:"resourceId"`
	RoleDefinition    pimRoleDefinition `json:"roleDefinition"`
	Resource          pimResource       `json:"resource"`
	AssignmentState   string            `json:"assignmentState"`
	MemberType        string            `json:"memberType"`
	EndDateTime       time.Time         `json:"endDateTime"`
	RequestedDateTime time.Time         `json:"requestedDateTime"`
	Reason            string            `json:"reason"`
	Status            any               `json:"status"`
}

type pimRoleAssignmentResp struct {
	Value []pimRoleAssignment `json:"value"`
}

// PIM API request structures for activation
type pimActivationSchedule struct {
	Type          string `json:"type"`
	StartDateTime any    `json:"startDateTime"`
	EndDateTime   any    `json:"endDateTime"`
	Duration      string `json:"duration"`
}

type pimActivationRequest struct {
	RoleDefinitionID string                `json:"roleDefinitionId"`
	ResourceID       string                `json:"resourceId"`
	SubjectID        string                `json:"subjectId"`
	AssignmentState  string                `json:"assignmentState"`
	Type             string                `json:"type"`
	Reason           string                `json:"reason"`
	Schedule         pimActivationSchedule `json:"schedule"`
}

type pimActivationResponse struct {
	Status struct {
		Status    string `json:"status"`
		SubStatus string `json:"subStatus"`
	} `json:"status"`
	RoleAssignmentEndDateTime time.Time `json:"roleAssignmentEndDateTime"`
}

// ListEligiblePIMGroups queries and displays all PIM groups the user is eligible for using Azure RBAC PIM API
func ListEligiblePIMGroups(ctx context.Context, cred azcore.TokenCredential, userID string) ([]pimRoleAssignment, error) {
	assignments, err := getRoleAssignments(ctx, cred, userID, "Eligible")
	if err != nil {
		return nil, err
	}

	return assignments, nil
}

// ListActivePIMGroups queries and displays all PIM groups the user has currently activated using Azure RBAC PIM API
func ListActivePIMGroups(ctx context.Context, cred azcore.TokenCredential, userID string) ([]pimRoleAssignment, error) {
	assignments, err := getRoleAssignments(ctx, cred, userID, "Active")
	if err != nil {
		return nil, err
	}

	return assignments, nil
}

// ListPendingPIMRequests queries and displays all pending PIM group activation requests for the user
func ListPendingPIMRequests(ctx context.Context, cred azcore.TokenCredential, userID string) ([]pimRoleAssignment, error) {
	assignments, err := getRoleAssignmentRequests(ctx, cred, userID, "PendingApproval")
	if err != nil {
		return nil, err
	}

	return assignments, nil
}

// RequestPIMGroupActivation requests activation for a PIM group using Azure RBAC PIM API
func RequestPIMGroupActivation(ctx context.Context, cred azcore.TokenCredential, userID,
	groupName, reason string, duration time.Duration, roleName string) (pimActivationResponse, error) {
	if roleName == "" {
		return pimActivationResponse{}, fmt.Errorf("role name must be specified")
	}

	if duration <= 0 {
		return pimActivationResponse{}, fmt.Errorf("duration must be greater than zero")
	}

	if groupName == "" {
		return pimActivationResponse{}, fmt.Errorf("group name must be specified")
	}

	// First, find the eligible role assignment for the specified group
	assignments, err := getRoleAssignments(ctx, cred, userID, "Eligible")
	if err != nil {
		return pimActivationResponse{}, err
	}

	var targetAssignment *pimRoleAssignment

	for _, assignment := range assignments {
		if assignment.Resource.DisplayName == groupName && strings.EqualFold(assignment.RoleDefinition.DisplayName, roleName) {
			targetAssignment = &assignment
			break
		}
	}

	if targetAssignment == nil {
		return pimActivationResponse{}, fmt.Errorf("no eligible group found: %s with role: %s", groupName, roleName)
	}

	// Convert duration to ISO 8601 duration format (e.g., PT720M for 720 minutes)
	durationMinutes := int(duration.Minutes())
	isoDuration := fmt.Sprintf("PT%dM", durationMinutes)

	if reason == "" {
		reason = "Requested via pim-cli"
	}

	// Prepare the request body
	requestBody := pimActivationRequest{
		RoleDefinitionID: targetAssignment.RoleDefinition.ID,
		ResourceID:       targetAssignment.ResourceID,
		SubjectID:        userID,
		AssignmentState:  "Active",
		Type:             "UserAdd",
		Reason:           reason,
		Schedule: pimActivationSchedule{
			Type:          "Once",
			StartDateTime: nil,
			EndDateTime:   nil,
			Duration:      isoDuration,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return pimActivationResponse{}, fmt.Errorf("failed to marshal activation request body: %w", err)
	}

	activationURL := fmt.Sprintf("%s/roleAssignmentRequests", pimAPIBaseURL)

	var response pimActivationResponse
	if err := pimAPIRequest(ctx, cred, http.MethodPost, activationURL, bodyBytes, &response); err != nil {
		return pimActivationResponse{}, err
	}

	return response, nil
}

// getRoleAssignments fetches role assignments for a user with the given filter
func getRoleAssignments(ctx context.Context, cred azcore.TokenCredential, userID, assignmentState string) ([]pimRoleAssignment, error) {
	filter := fmt.Sprintf("subjectId eq '%s'", userID)
	if assignmentState != "" {
		filter += fmt.Sprintf(" and assignmentState eq '%s'", assignmentState)
	}

	reqURL := fmt.Sprintf("%s/roleAssignments?$filter=%s&$expand=resource,roleDefinition",
		pimAPIBaseURL, url.QueryEscape(filter))

	var pimResp pimRoleAssignmentResp
	if err := pimAPIRequest(ctx, cred, http.MethodGet, reqURL, nil, &pimResp); err != nil {
		return nil, err
	}

	return pimResp.Value, nil
}

// getRoleAssignmentRequests fetches role assignment requests for a user with the given status filter
func getRoleAssignmentRequests(ctx context.Context, cred azcore.TokenCredential, userID, status string) ([]pimRoleAssignment, error) {
	filter := fmt.Sprintf("subjectId eq '%s'", userID)
	if status != "" {
		filter += fmt.Sprintf(" and status/subStatus eq '%s'", status)
	}

	reqURL := fmt.Sprintf("%s/roleAssignmentRequests?$filter=%s&$expand=resource,roleDefinition",
		pimAPIBaseURL, url.QueryEscape(filter))

	var pimResp pimRoleAssignmentResp
	if err := pimAPIRequest(ctx, cred, http.MethodGet, reqURL, nil, &pimResp); err != nil {
		return nil, err
	}

	return pimResp.Value, nil
}

// pimAPIRequest performs an authenticated request to the PIM API and decodes the response
func pimAPIRequest(ctx context.Context, cred azcore.TokenCredential, method, url string, body []byte, result any) error {
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{pimAPIScope},
	})
	if err != nil {
		return fmt.Errorf("failed to get PIM API token: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query PIM API: %w", err)
	}

	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBodyStr := strings.TrimSpace(string(respBody))
		return fmt.Errorf("PIM API error: %s - %s", resp.Status, respBodyStr)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
