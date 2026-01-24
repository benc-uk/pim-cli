package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/benc-uk/pimg-cli/pkg/graph"
	"github.com/spf13/cobra"
)

var userID string
var userName string
var tenantName string

var rootCmd = &cobra.Command{
	Use:   "pim-cli",
	Short: "PIM Group Management CLI",
	Long:  `A command-line tool to manage access to Privileged Identity Management (PIM) groups in Azure.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		_ = cmd.Help()
	},
}

// Execute executes the root command.
func Execute(ver string) error {
	fmt.Printf("PIM Group Management CLI v%s\n", ver)
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(activeCmd)
	rootCmd.AddCommand(requestCmd)
}

// authenticate creates Azure credential and Microsoft Graph client
func authenticate() (azcore.TokenCredential, *graph.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Note. getting here does not guarantee that authentication will succeed!

	// Create Graph client using HTTP-based implementation
	graphClient := graph.NewClient(cred)

	return cred, graphClient, nil
}

func getUserTenantInfo(graphClient *graph.Client) {
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
