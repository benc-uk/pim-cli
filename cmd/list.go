package cmd

import (
	"context"
	"fmt"
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
			assignments, err := pim.ListEligiblePIMGroups(ctx, cred, user.ID)
			if err != nil {
				log.Fatalf("Failed to list eligible PIM groups: %v", err)
			}

			if len(assignments) == 0 {
				fmt.Println("\nNo eligible PIM groups found")
			}

			printer(fmt.Sprintf("\nFound %d eligible PIM group(s):\n", len(assignments)))

			for _, assignment := range assignments {
				fmt.Printf("%s\n", assignment.Resource.DisplayName)
				printer(fmt.Sprintf("  Role: %s", assignment.RoleDefinition.DisplayName))
				printer(fmt.Sprintf("  Member Type: %s", assignment.MemberType))
				printer(fmt.Sprintf("  Resource ID: %s\n", assignment.ResourceID))
			}
		}
	},
}

func init() {
	listCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "List all Entra ID groups, not just eligible PIM ones")
}
