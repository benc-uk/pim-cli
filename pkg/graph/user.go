package graph

import (
	"context"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/organization"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/users"
)

// GetUserInfo gets the current user's object ID and display name using Microsoft Graph SDK
func GetUserInfo(ctx context.Context, graphClient *msgraphsdk.GraphServiceClient) (string, string, error) {
	requestParams := &users.UserItemRequestBuilderGetQueryParameters{
		Select: []string{"id", "displayName"},
	}
	config := &users.UserItemRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParams,
	}

	user, err := graphClient.Me().Get(ctx, config)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user info: %w", err)
	}

	userID := ""
	if user.GetId() != nil {
		userID = *user.GetId()
	}

	displayName := ""
	if user.GetDisplayName() != nil {
		displayName = *user.GetDisplayName()
	}

	return userID, displayName, nil
}

// GetTenantInfo gets the current tenant's display name using Microsoft Graph SDK
func GetTenantInfo(ctx context.Context, graphClient *msgraphsdk.GraphServiceClient) (string, error) {
	requestParams := &organization.OrganizationRequestBuilderGetQueryParameters{
		Select: []string{"displayName"},
	}
	config := &organization.OrganizationRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParams,
	}

	result, err := graphClient.Organization().Get(ctx, config)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant info: %w", err)
	}

	orgs := result.GetValue()
	if len(orgs) == 0 {
		return "", fmt.Errorf("no organization found")
	}

	displayName := ""
	if orgs[0].GetDisplayName() != nil {
		displayName = *orgs[0].GetDisplayName()
	}

	return displayName, nil
}
