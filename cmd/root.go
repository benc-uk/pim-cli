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

var user graph.User
var tenantName string
var quietMode bool
var version string

var rootCmd = &cobra.Command{
	Use:   "pim-cli",
	Short: "PIM Group Management CLI",
	Long:  `A command-line tool to manage access to Privileged Identity Management (PIM) groups in Azure.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This runs after flag parsing, so quietMode is available
		printer(fmt.Sprintf("PIM Group CLI v%s", version))
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		_ = cmd.Help()
	},
}

// Execute executes the root command.
func Execute(ver string) error {
	version = ver

	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(activeCmd)
	rootCmd.AddCommand(requestCmd)
	// rootCmd.AddCommand(extendCmd) // This command didn't really work, removed maybe to re-add later

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "Suppress most output")
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

// getUserTenantInfo retrieves and displays the current user and tenant information
func getUserTenantInfo(graphClient *graph.Client) {
	ctx := context.Background()

	// Get current user info
	var err error

	user, err = graph.GetCurrentUser(ctx, graphClient)
	if err != nil {
		log.Fatalf("Failed to get user info: %v", err)
	}

	if !quietMode {
		// Get tenant info, micro speed up by only doing this if not quiet mode
		tenantName, err = graph.GetTenantInfo(ctx, graphClient)
		if err != nil {
			log.Fatalf("Failed to get tenant info: %v", err)
		}

		printer(fmt.Sprintf("Tenant: %s", tenantName))
		printer(fmt.Sprintf("Current user: %s", user.DisplayName))
	}
}

func printer(msg string) {
	if !quietMode {
		fmt.Println(msg)
	}
}
