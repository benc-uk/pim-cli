package graph

import (
	"context"
	"fmt"
	"sort"

	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
)

// ListAllGroups lists all AAD groups using Microsoft Graph SDK with pagination
func ListAllGroups(ctx context.Context, graphClient *msgraphsdk.GraphServiceClient) error {
	requestParams := &groups.GroupsRequestBuilderGetQueryParameters{
		Select:  []string{"id", "displayName", "description", "groupTypes"},
		Orderby: []string{"displayName"},
	}
	config := &groups.GroupsRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParams,
	}

	// Collect all groups across pages
	var allGroups []models.Groupable

	fmt.Println("Fetching all groups from EntraID...")
	result, err := graphClient.Groups().Get(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to list groups: %w", err)
	}

	allGroups = append(allGroups, result.GetValue()...)

	// Follow pagination links
	for result.GetOdataNextLink() != nil && *result.GetOdataNextLink() != "" {
		result, err = graphClient.Groups().WithUrl(*result.GetOdataNextLink()).Get(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to get next page of groups: %w", err)
		}
		allGroups = append(allGroups, result.GetValue()...)
	}

	if len(allGroups) == 0 {
		fmt.Println("\nNo groups found")
		return nil
	}

	sort.SliceStable(allGroups, func(i, j int) bool {
		nameI := ""
		if allGroups[i].GetDisplayName() != nil {
			nameI = *allGroups[i].GetDisplayName()
		}
		nameJ := ""
		if allGroups[j].GetDisplayName() != nil {
			nameJ = *allGroups[j].GetDisplayName()
		}
		return nameI < nameJ
	})

	fmt.Printf("\nFound %d group(s):\n\n", len(allGroups))
	for _, group := range allGroups {
		name := "Unknown"
		if group.GetDisplayName() != nil {
			name = *group.GetDisplayName()
		}

		description := ""
		if group.GetDescription() != nil {
			description = *group.GetDescription()
		}

		groupID := ""
		if group.GetId() != nil {
			groupID = *group.GetId()
		}

		fmt.Printf("%s\n", name)
		if description != "" {
			fmt.Printf("  Description: %s\n", description)
		}
		fmt.Printf("  ID: %s\n", groupID)
		fmt.Println()
	}

	return nil
}
