package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/benc-uk/pim-cli/pkg/graph"
	"github.com/benc-uk/pim-cli/pkg/output"
	"github.com/benc-uk/pim-cli/pkg/pim"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var allFlag bool

// roleInfo holds role and member type for a group assignment
type roleInfo struct {
	role       string
	memberType string
}

// groupInfo holds condensed information about a group with multiple roles
type groupInfo struct {
	name  string
	roles []roleInfo
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List eligible groups",
	Long:  `List all eligible groups for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		cred, graphClient, err := getCredentials()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		getUserTenantInfo(graphClient)
		ctx := context.Background()

		if allFlag {
			err = graph.ListAllGroups(ctx, graphClient)
			if err != nil {
				output.Fatalf("Failed to list all groups: %v", err)
			}
		} else {
			assignments, err := pim.ListEligiblePIMGroups(ctx, cred, user.ID)
			if err != nil {
				output.Fatalf("Failed to list eligible PIM groups: %v", err)
			}

			if len(assignments) == 0 {
				output.Printfq("No eligible PIM groups found\n")
				return
			}

			// Condense assignments by group name
			groupMap := make(map[string]*groupInfo)
			groupOrder := []string{} // Preserve order

			for _, assignment := range assignments {
				name := assignment.Resource.DisplayName
				if _, exists := groupMap[name]; !exists {
					groupMap[name] = &groupInfo{name: name}
					groupOrder = append(groupOrder, name)
				}
				groupMap[name].roles = append(groupMap[name].roles, roleInfo{
					role:       assignment.RoleDefinition.DisplayName,
					memberType: assignment.MemberType,
				})
			}

			output.Printf("Found %d eligible PIM group(s):\n\n", len(groupMap))

			var tbl table.Table
			if quietMode {
				tbl = table.New("Group Name", "Roles")
				tbl.WithHeaderFormatter(func(format string, a ...interface{}) string {
					return fmt.Sprintf("\033[33m"+format+"\033[0m", a...) // Bold
				})
			}

			for _, name := range groupOrder {
				info := groupMap[name]

				if quietMode {
					roleNames := make([]string, len(info.roles))
					for i, r := range info.roles {
						roleNames[i] = r.role
					}
					tbl.AddRow(info.name, strings.Join(roleNames, ", "))
					continue
				}

				output.Printf("\033[33m%s\033[0m\n", info.name)
				for _, r := range info.roles {
					output.Printf("  \033[34mRole:\033[0m\t\t%s (%s)\n", r.role, r.memberType)
				}
				output.Println()
			}

			if quietMode {
				tbl.Print()
			}
		}
	},
}

func init() {
	listCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "List all Entra ID groups, not just eligible PIM ones")
}
