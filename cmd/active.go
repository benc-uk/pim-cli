package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/benc-uk/pim-cli/pkg/output"
	"github.com/benc-uk/pim-cli/pkg/pim"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "List active group activations",
	Long:  `List all active PIM group + role activations for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		cred, graphClient, err := getCredentials()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		getUserTenantInfo(graphClient)
		ctx := context.Background()

		assignments, err := pim.ListActivePIMGroups(ctx, cred, user.ID)
		if err != nil {
			output.Fatalf("Failed to list active groups: %v\n", err)
		}

		if len(assignments) == 0 {
			output.Printfq("No active groups found\n")
			return
		}

		output.Printf("Found %d active group(s):\n\n", len(assignments))
		// Find the longest group name for alignment
		maxNameLen := 0
		for _, assignment := range assignments {
			if len(assignment.Resource.DisplayName) > maxNameLen {
				maxNameLen = len(assignment.Resource.DisplayName)
			}
		}

		var tbl table.Table
		if quietMode {
			tbl = table.New("Group Name", "Role", "Expires", "Time Left")
			tbl.WithHeaderFormatter(func(format string, a ...interface{}) string {
				return fmt.Sprintf("\033[33m"+format+"\033[0m", a...) // Bold
			})
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

			// The API returns status as a mix of types, so we need to assert it down to something useful
			status := "Unknown"
			if assignment.Status != nil {
				if statusStr, ok := assignment.Status.(string); ok {
					status = statusStr
				}
			}

			if quietMode {
				tbl.AddRow(assignment.Resource.DisplayName, assignment.RoleDefinition.DisplayName, expiresNice, leftNice)
				continue
			}

			output.Printf("\033[33m%s\033[0m\n", assignment.Resource.DisplayName)
			output.Printf("  \033[34mRole:\033[0m\t\t%s\n", assignment.RoleDefinition.DisplayName)
			output.Printf("  \033[34mMember Type:\033[0m\t%s\n", assignment.MemberType)
			output.Printf("  \033[34mExpires:\033[0m\t%s \033[36m(%s)\033[0m\n", expiresNice, leftNice)
			output.Printf("  \033[34mStatus:\033[0m\t%s\n\n", status)
		}

		if quietMode {
			tbl.Print()
		}
	},
}
