package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/benc-uk/pimg-cli/pkg/graph"
	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
	"github.com/spf13/cobra"
)

var userID string
var userName string
var tenantName string

var rootCmd = &cobra.Command{
	Use:   "pimg-cli",
	Short: "PIM Group Management CLI",
	Long:  `A command-line tool to manage access to Privileged Identity Management (PIM) groups in Azure.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		cmd.Help()
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(activeCmd)
	rootCmd.AddCommand(requestCmd)
}

// authenticate creates Azure credential and Microsoft Graph client
func authenticate() (azcore.TokenCredential, *msgraphsdk.GraphServiceClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Using the beta SDK with default Graph scopes
	graphClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	return cred, graphClient, nil
}

func getUserTenantInfo(graphClient *msgraphsdk.GraphServiceClient) {
	fmt.Println("Successfully authenticated using Azure CLI credentials")

	ctx := context.Background()

	// Get current user info
	var err error
	userID, userName, err = graph.GetUserInfo(ctx, graphClient)
	if err != nil {
		log.Fatalf("Failed to get user info: %v", err)
	}

	// Get tenant info
	tenantName, err = graph.GetTenantInfo(ctx, graphClient)
	if err != nil {
		log.Fatalf("Failed to get tenant info: %v", err)
	}

	fmt.Printf("Tenant: %s\n", tenantName)
	fmt.Printf("Current user: %s\n", userName)
}
