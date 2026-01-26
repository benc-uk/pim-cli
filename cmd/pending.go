package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/benc-uk/pimg-cli/pkg/pim"
	"github.com/spf13/cobra"
)

var pendingCmd = &cobra.Command{
	Use:     "pending",
	Short:   "List pending PIM groups",
	Aliases: []string{"status"},
	Long:    `List all pending PIM groups for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		cred, graphClient, err := authenticate()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		getUserTenantInfo(graphClient)
		ctx := context.Background()

		pendingAssignments, err := pim.ListPendingPIMRequests(ctx, cred, user.ID)
		if err != nil {
			log.Fatalf("Failed to list pending PIM group requests: %v", err)
		}

		if len(pendingAssignments) == 0 {
			fmt.Println("No pending PIM group requests found")
			return
		}

		printer(fmt.Sprintf("\nFound %d pending PIM group request(s):\n", len(pendingAssignments)))

		// Find the longest group name for alignment
		maxPendingNameLen := 0
		for _, assignment := range pendingAssignments {
			if len(assignment.Resource.DisplayName) > maxPendingNameLen {
				maxPendingNameLen = len(assignment.Resource.DisplayName)
			}
		}

		if quietMode {
			fmt.Printf("\n%-*s %s\t\t%s\n", maxPendingNameLen, "Group Name", "Requested At", "Status")
		}

		for _, assignment := range pendingAssignments {
			requestedAtNice := assignment.RequestedDateTime.Format("15:04, Jan 02")

			// The API returns status as a mix of types, so we need to assert it down to something useful
			status := "Unknown"
			if assignment.Status != nil {
				if statusMap, ok := assignment.Status.(map[string]interface{}); ok {
					status = fmt.Sprintf("%s %s", statusMap["status"], statusMap["subStatus"])
				}
			}

			if quietMode {
				fmt.Printf("%-*s %s\t\t%s\n", maxPendingNameLen, assignment.Resource.DisplayName, requestedAtNice, status)
				continue
			}

			fmt.Printf("%s\n", assignment.Resource.DisplayName)
			fmt.Printf("  Role: %s\n", assignment.RoleDefinition.DisplayName)
			fmt.Printf("  Member Type: %s\n", assignment.MemberType)
			fmt.Printf("  Resource ID: %s\n", assignment.ResourceID)
			fmt.Printf("  Requested At: %s\n", requestedAtNice)
			fmt.Printf("  Status: %s\n\n", status)
		}
	},
}
