package cmd

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/benc-uk/pimg-cli/pkg/output"
	"github.com/benc-uk/pimg-cli/pkg/pim"
	"github.com/spf13/cobra"
)

var nameFlag string
var reasonFlag string
var durationFlag time.Duration
var roleFlag string

var requestCmd = &cobra.Command{
	Use:     "request",
	Short:   "Request activation for a group & role",
	Aliases: []string{"activate"},
	Long:    `Request activation for an eligible PIM group with the specified role for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		cred, graphClient, err := getCredentials()
		if err != nil {
			output.Fatalf("Authentication failed: %v\n", err)
		}

		getUserTenantInfo(graphClient)

		output.Printfq("Requesting '%s' role for '%s'...\n", roleFlag, nameFlag)
		response, err := pim.RequestPIMGroupActivation(ctx, cred, user.ID, nameFlag, reasonFlag, durationFlag, roleFlag)
		if err != nil {
			if strings.Contains(err.Error(), "RoleAssignmentExists") {
				output.Printfq("An active or pending assignment already exists for '%s'\n", nameFlag)
				os.Exit(0)
			}

			output.Fatalf("Failed to request PIM group activation: %v\n", err)
		}

		if response.Status.Status != "" {
			output.Printfq("Success. Status: %s\n", response.Status.Status)
		} else {
			output.Printfq("Activation request submitted. Response:\n %+v", response)
		}
	},
}

func init() {
	requestCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Name of the PIM group to request activation for (required)")
	requestCmd.Flags().StringVarP(&reasonFlag, "reason", "r", "", "Reason for requesting activation (required)")
	requestCmd.Flags().StringVarP(&roleFlag, "role", "o", "Member", "Role name to activate (e.g., 'Member', 'Owner')")
	requestCmd.Flags().DurationVarP(&durationFlag, "duration", "d", 12*time.Hour, "Duration for the activation (e.g., 30m, 1h, 2h)")

	_ = requestCmd.MarkFlagRequired("name")
	_ = requestCmd.MarkFlagRequired("reason")
}
