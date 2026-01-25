// ==============================================================================================
// Lightweight Microsoft Graph API client wrapper
//
// groups.go: Functions for interacting with Entra ID groups
// ===============================================================================================

package graph

import (
	"context"
	"fmt"
	"net/http"
	"sort"
)

// Group represents an Azure AD group from the Graph API
type Group struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description"`
	GroupTypes  []string `json:"groupTypes"`
}

// groupsResponse represents the paginated response from the Graph API
type groupsResponse struct {
	Value    []Group `json:"value"`
	NextLink string  `json:"@odata.nextLink"`
}

// ListAllGroups lists all AAD groups using Microsoft Graph REST API with pagination
func ListAllGroups(ctx context.Context, client *Client) error {
	// Collect all groups across pages
	var allGroups []Group

	fmt.Println("Fetching all groups from EntraID...")

	reqURL := graphAPIBaseURL + "/groups?$select=id,displayName,description,groupTypes&$orderby=displayName"

	for reqURL != "" {
		var resp groupsResponse
		if err := client.Request(ctx, http.MethodGet, reqURL, nil, &resp); err != nil {
			return fmt.Errorf("failed to list groups: %w", err)
		}

		allGroups = append(allGroups, resp.Value...)
		reqURL = resp.NextLink
	}

	if len(allGroups) == 0 {
		fmt.Println("\nNo groups found")
		return nil
	}

	sort.SliceStable(allGroups, func(i, j int) bool {
		return allGroups[i].DisplayName < allGroups[j].DisplayName
	})

	fmt.Printf("\nFound %d group(s):\n\n", len(allGroups))

	for _, group := range allGroups {
		name := group.DisplayName
		if name == "" {
			name = "Unknown"
		}

		fmt.Printf("%s\n", name)

		if group.Description != "" {
			fmt.Printf("  Description: %s\n", group.Description)
		}

		fmt.Printf("  ID: %s\n", group.ID)
		fmt.Println()
	}

	return nil
}
