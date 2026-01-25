// ==============================================================================================
// Lightweight Microsoft Graph API client wrapper
//
// user.go: Functions for interacting with Entra ID users & tenants
// ===============================================================================================

package graph

import (
	"context"
	"fmt"
	"net/http"
)

// User represents the current user from the Graph API
type User struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	JobTitle          string `json:"jobTitle"`
	Department        string `json:"department"`
	OfficeLocation    string `json:"officeLocation"`
	City              string `json:"city"`
	CompanyName       string `json:"companyName"`
	EmployeeID        string `json:"employeeId"`
	AccountEnabled    bool   `json:"accountEnabled"`
	Alias             string `json:"onPremisesSamAccountName"`
}

// Organization represents a tenant organization from the Graph API
type Organization struct {
	DisplayName string `json:"displayName"`
}

// organizationResponse represents the response from the organization endpoint
type organizationResponse struct {
	Value []Organization `json:"value"`
}

// GetCurrentUser gets the current user's object ID and display name using Microsoft Graph REST API
func GetCurrentUser(ctx context.Context, client *Client) (User, error) {
	reqURL := graphAPIBaseURL + "/me"

	var user User
	if err := client.Request(ctx, http.MethodGet, reqURL, nil, &user); err != nil {
		return User{}, fmt.Errorf("failed to get user info: %w", err)
	}

	return user, nil
}

// GetTenantInfo gets the current tenant's display name using Microsoft Graph REST API
func GetTenantInfo(ctx context.Context, client *Client) (string, error) {
	reqURL := graphAPIBaseURL + "/organization?$select=displayName"

	var resp organizationResponse
	if err := client.Request(ctx, http.MethodGet, reqURL, nil, &resp); err != nil {
		return "", fmt.Errorf("failed to get tenant info: %w", err)
	}

	if len(resp.Value) == 0 {
		return "", fmt.Errorf("no organization found")
	}

	return resp.Value[0].DisplayName, nil
}
