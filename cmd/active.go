package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/benc-uk/pimg-cli/pkg/pim"
	"github.com/spf13/cobra"
)

var activeCmd = &cobra.Command{
	Use:     "active",
	Short:   "List active PIM groups",
	Aliases: []string{"status"},
	Long:    `List all active PIM groups for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		cred, graphClient, err := authenticate()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		getUserTenantInfo(graphClient)
		ctx := context.Background()

		assignments, err := pim.ListActivePIMGroups(ctx, cred, user.ID)
		if err != nil {
			log.Fatalf("Failed to list active PIM groups: %v", err)
		}

		if len(assignments) == 0 {
			fmt.Println("\nNo active PIM groups found")
		}

		printer(fmt.Sprintf("\nFound %d active PIM group(s):\n", len(assignments)))

		// Find the longest group name for alignment
		maxNameLen := 0
		for _, assignment := range assignments {
			if len(assignment.Resource.DisplayName) > maxNameLen {
				maxNameLen = len(assignment.Resource.DisplayName)
			}
		}

		if quietMode {
			fmt.Printf("%-*s %s\t\t%s\n", maxNameLen, "Group Name", "Expires", "Left")
		}

		for _, assignment := range assignments {
			expiresNice := assignment.EndDateTime.Format("15:04, Jan 02")
			remaining := time.Until(assignment.EndDateTime)
			remaining = remaining.Round(time.Minute)
			h := remaining / time.Hour
			remaining -= h * time.Hour
			m := remaining / time.Minute
			leftNice := fmt.Sprintf("%dh %dm", h, m)

			if assignment.EndDateTime.IsZero() {
				expiresNice = "Never expires"
				leftNice = "N/A"
			}

			if quietMode {
				fmt.Printf("%-*s %s\t%s\n", maxNameLen, assignment.Resource.DisplayName, expiresNice, leftNice)
				continue
			}

			fmt.Printf("%s\n", assignment.Resource.DisplayName)
			fmt.Printf("  Role: %s\n", assignment.RoleDefinition.DisplayName)
			fmt.Printf("  Member Type: %s\n", assignment.MemberType)
			fmt.Printf("  Resource ID: %s\n", assignment.ResourceID)
			fmt.Printf("  Expires: %s (%s)\n\n", expiresNice, leftNice)
		}
	},
}
