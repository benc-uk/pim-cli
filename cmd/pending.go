package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/benc-uk/pimg-cli/pkg/output"
	"github.com/benc-uk/pimg-cli/pkg/pim"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var pendingCmd = &cobra.Command{
	Use:     "pending",
	Short:   "List pending requests",
	Aliases: []string{"status"},
	Long:    `List all pending PIM group + role activation requests for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		cred, graphClient, err := getCredentials()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		getUserTenantInfo(graphClient)
		ctx := context.Background()

		pendingAssignments, err := pim.ListPendingPIMRequests(ctx, cred, user.ID)
		if err != nil {
			output.Fatalf("Failed to list pending requests: %v\n", err)
		}

		if len(pendingAssignments) == 0 {
			output.Printfq("No pending requests found\n")
			return
		}

		output.Printf("Found %d pending request(s):\n\n", len(pendingAssignments))

		var tbl table.Table
		if quietMode {
			tbl = table.New("Group Name", "Role", "Requested At", "Status")
			tbl.WithHeaderFormatter(func(format string, a ...interface{}) string {
				return fmt.Sprintf("\033[33m"+format+"\033[0m", a...) // Bold
			})
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
				tbl.AddRow(assignment.Resource.DisplayName, assignment.RoleDefinition.DisplayName, requestedAtNice, status)
				continue
			}

			output.Printf("\033[33m%s\033[0m\n", assignment.Resource.DisplayName)
			output.Printf("  \033[34mRole:\033[0m\t\t%s\n", assignment.RoleDefinition.DisplayName)
			output.Printf("  \033[34mRequested At:\033[0m\t%s\n", requestedAtNice)
			output.Printf("  \033[34mStatus:\033[0m\t%s\n\n", status)
		}

		if quietMode {
			tbl.Print()
		}
	},
}
