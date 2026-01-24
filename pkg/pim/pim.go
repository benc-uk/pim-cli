// ===========================================================================================
// Provides functions to interact with Azure RBAC PIM API
// for managing Privileged Identity Management (PIM) group activations.
//
// Why not use Microsoft Graph API? Because it requires a permissions
// not available via Azure CLI authentication (e.g. PrivilegedAccess.ReadWrite.AzureADGroup).
// So we use the PIM API directly instead!
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
	ID              string            `json:"id"`
	ResourceID      string            `json:"resourceId"`
	RoleDefinition  pimRoleDefinition `json:"roleDefinition"`
	Resource        pimResource       `json:"resource"`
	AssignmentState string            `json:"assignmentState"`
	MemberType      string            `json:"memberType"`
	EndDateTime     time.Time         `json:"endDateTime"`
}

type pimResponse struct {
	Value []pimRoleAssignment `json:"value"`
}

// ListEligiblePIMGroups queries and displays all PIM groups the user is eligible for using Azure RBAC PIM API
func ListEligiblePIMGroups(ctx context.Context, cred azcore.TokenCredential, userID string) error {
	fmt.Println("Fetching all eligible PIM groups from EntraID...")

	assignments, err := getRoleAssignments(ctx, cred, userID, "Eligible")
	if err != nil {
		return err
	}

	if len(assignments) == 0 {
		fmt.Println("\nNo eligible PIM groups found")
		return nil
	}

	fmt.Printf("\nFound %d eligible PIM group(s):\n\n", len(assignments))
	for _, assignment := range assignments {
		fmt.Printf("%s\n", assignment.Resource.DisplayName)
		fmt.Printf("  Role: %s\n", assignment.RoleDefinition.DisplayName)
		fmt.Printf("  Member Type: %s\n", assignment.MemberType)
		fmt.Printf("  Resource ID: %s\n", assignment.ResourceID)
		fmt.Println()
	}

	return nil
}

// ListActivePIMGroups queries and displays all PIM groups the user has currently activated using Azure RBAC PIM API
func ListActivePIMGroups(ctx context.Context, cred azcore.TokenCredential, userID string) error {
	fmt.Println("Fetching all active PIM groups from EntraID...")

	assignments, err := getRoleAssignments(ctx, cred, userID, "Active")
	if err != nil {
		return err
	}

	if len(assignments) == 0 {
		fmt.Println("\nNo active PIM groups found")
		return nil
	}

	fmt.Printf("\nFound %d active PIM group(s):\n\n", len(assignments))
	for _, assignment := range assignments {
		expiresNice := assignment.EndDateTime.Format("15:04, Jan 02")
		remaining := time.Until(assignment.EndDateTime)
		remaining = remaining.Round(time.Minute)
		h := remaining / time.Hour
		remaining -= h * time.Hour
		m := remaining / time.Minute
		leftNice := fmt.Sprintf("%dh %dm", h, m)

		if assignment.EndDateTime.IsZero() {
			expiresNice = "Never"
			leftNice = "N/A"
		}

		fmt.Printf("%s\n", assignment.Resource.DisplayName)
		fmt.Printf("  Role: %s\n", assignment.RoleDefinition.DisplayName)
		fmt.Printf("  Member Type: %s\n", assignment.MemberType)
		fmt.Printf("  Resource ID: %s\n", assignment.ResourceID)
		fmt.Printf("  Expires: %s (%s)\n", expiresNice, leftNice)
		fmt.Println()
	}

	return nil
}

// RequestPIMGroupActivation requests activation for a PIM group using Azure RBAC PIM API
func RequestPIMGroupActivation(ctx context.Context, cred azcore.TokenCredential, userID, groupName, reason string, duration time.Duration) error {
	// First, find the eligible role assignment for the specified group
	assignments, err := getRoleAssignments(ctx, cred, userID, "Eligible")
	if err != nil {
		return err
	}

	var targetAssignment *pimRoleAssignment
	for _, assignment := range assignments {
		if assignment.Resource.DisplayName == groupName {
			targetAssignment = &assignment
			break
		}
	}

	if targetAssignment == nil {
		return fmt.Errorf("no eligible PIM group found with name: %s", groupName)
	}

	// Convert duration to ISO 8601 duration format (e.g., PT720M for 720 minutes)
	durationMinutes := int(duration.Minutes())
	isoDuration := fmt.Sprintf("PT%dM", durationMinutes)

	if reason == "" {
		reason = "Requested via pimg-cli"
	}

	// Prepare the request body matching the shell script format
	requestBody := map[string]interface{}{
		"roleDefinitionId": targetAssignment.RoleDefinition.ID,
		"resourceId":       targetAssignment.ResourceID,
		"subjectId":        userID,
		"assignmentState":  "Active",
		"type":             "UserAdd",
		"reason":           reason,
		"ticketNumber":     "",
		"ticketSystem":     "",
		"schedule": map[string]interface{}{
			"type":          "Once",
			"startDateTime": nil,
			"endDateTime":   nil,
			"duration":      isoDuration,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal activation request body: %w", err)
	}

	fmt.Printf("Requesting activation for PIM group '%s'...\n", groupName)

	activationURL := fmt.Sprintf("%s/roleAssignmentRequests", pimAPIBaseURL)

	var response struct {
		Status struct {
			Status    string `json:"status"`
			SubStatus string `json:"subStatus"`
		} `json:"status"`
	}

	if err := pimAPIRequest(ctx, cred, http.MethodPost, activationURL, bodyBytes, &response); err != nil {
		return err
	}

	if response.Status.Status != "" {
		fmt.Printf("Successfully requested activation for PIM group '%s'\n", groupName)
		fmt.Printf("Status: %s", response.Status.Status)
		if response.Status.SubStatus != "" {
			fmt.Printf(" (%s)", response.Status.SubStatus)
		}
		fmt.Println()
	} else {
		fmt.Printf("Successfully requested activation for PIM group '%s' for %d minutes\n", groupName, durationMinutes)
	}

	return nil
}

// getRoleAssignments fetches role assignments for a user with the given filter
func getRoleAssignments(ctx context.Context, cred azcore.TokenCredential, userID, assignmentState string) ([]pimRoleAssignment, error) {
	filter := fmt.Sprintf("subjectId eq '%s'", userID)
	if assignmentState != "" {
		filter += fmt.Sprintf(" and assignmentState eq '%s'", assignmentState)
	}

	reqURL := fmt.Sprintf("%s/roleAssignments?$filter=%s&$expand=resource,roleDefinition", pimAPIBaseURL, url.QueryEscape(filter))

	var pimResp pimResponse
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
		return fmt.Errorf("PIM API error: %s - %s", resp.Status, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
