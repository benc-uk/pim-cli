package cmd

import (
	"context"
	"log"

	"github.com/benc-uk/pimg-cli/pkg/graph"
	"github.com/benc-uk/pimg-cli/pkg/pim"
	"github.com/spf13/cobra"
)

var allFlag bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List eligible PIM groups",
	Long:  `List all eligible PIM groups for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		cred, graphClient, err := authenticate()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		getUserTenantInfo(graphClient)
		ctx := context.Background()

		if allFlag {
			err = graph.ListAllGroups(ctx, graphClient)
			if err != nil {
				log.Fatalf("Failed to list all groups: %v", err)
			}
		} else {
			err = pim.ListEligiblePIMGroups(ctx, cred, userID)
			if err != nil {
				log.Fatalf("Failed to list eligible PIM groups: %v", err)
			}
		}
	},
}

func init() {
	listCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "List all Entra ID groups, not just eligible PIM ones")
}
