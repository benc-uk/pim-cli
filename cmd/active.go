package cmd

import (
	"context"
	"log"

	"github.com/benc-uk/pimg-cli/pkg/pim"
	"github.com/spf13/cobra"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "List active PIM groups",
	Long:  `List all active PIM groups for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		cred, graphClient, err := authenticate()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		getUserTenantInfo(graphClient)
		ctx := context.Background()

		err = pim.ListActivePIMGroups(ctx, cred, userID)
		if err != nil {
			log.Fatalf("Failed to list active PIM groups: %v", err)
		}
	},
}
