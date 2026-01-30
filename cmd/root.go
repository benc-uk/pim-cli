package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/benc-uk/pimg-cli/pkg/graph"
	"github.com/benc-uk/pimg-cli/pkg/output"
	"github.com/spf13/cobra"
)

var user graph.User
var tenantName string
var quietMode bool
var version string

var rootCmd = &cobra.Command{
	Use:   "pim-cli",
	Short: "PIM Group Management CLI",
	Long:  `A command-line tool to manage access to Privileged Identity Management (PIM) groups in Azure`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This runs after flag parsing, so quietMode is available
		if quietMode {
			output.SetLevel(output.Quiet)
		} else {
			output.SetLevel(output.Normal)
		}

		output.Printf("\033[35mPIM Group CLI v%s\033[0m\n", version)
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
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(pendingCmd)
	rootCmd.AddCommand(requestCmd)

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "Simple output in tabular format")
}

// getCredentials creates Azure credential and Microsoft Graph client
func getCredentials() (azcore.TokenCredential, *graph.Client, error) {
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

	// Get current user info, the user details are used in multiple commands
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
	}

	output.Printf("\033[34mTenant:\033[0m\t\t%s\n", tenantName)
	output.Printf("\033[34mCurrent user:\033[0m\t%s\n", user.DisplayName)
}
