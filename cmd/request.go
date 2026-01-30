package cmd

import (
	"context"
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
		cred, graphClient, err := getCredentials()
		if err != nil {
			output.Fatalf("Authentication failed: %v\n", err)
		}

		getUserTenantInfo(graphClient)
		ctx := context.Background()

		output.Printfq("Requesting '\033[1;32m%s\033[0m' role for '\033[1;32m%s\033[0m'...\n", roleFlag, nameFlag)
		response, err := pim.RequestPIMGroupActivation(ctx, cred, user.ID, nameFlag, reasonFlag, durationFlag, roleFlag)
		status := strings.TrimSpace(response.Status.Status)
		if err != nil {
			// Check for http 400, and don't treat as fatal, as it's likely a role already active, which is cool
			if pimErr, ok := err.(*pim.PimError); ok {
				if pimErr.HTTPStatusCode == 400 {
					status = pimErr.ApiError.Message
				} else {
					output.Fatalf("Activation failed: %v\n", err)
				}
			} else {
				output.Fatalf("%v\n", err)
			}
		}

		if status != "" {
			output.Printfq("\033[34mRequest:\033[0m %s\n", status)
		} else {
			// Unexpected response format, print full response, you should not normally see this
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
