package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/benc-uk/pimg-cli/pkg/pim"
	"github.com/spf13/cobra"
)

var nameFlag string
var reasonFlag string
var durationFlag time.Duration

var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request activation for a PIM group",
	Long:  `Request activation for an eligible PIM group for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		cred, graphClient, err := authenticate()
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		getUserTenantInfo(graphClient)
		if reasonFlag == "" {
			reasonFlag = "Standard activation request via pimg-cli for " + user.DisplayName
		}

		printer(fmt.Sprintf("Requesting activation for PIM group '%s'...", nameFlag))
		response, err := pim.RequestPIMGroupActivation(ctx, cred, user.ID, nameFlag, reasonFlag, durationFlag)
		if err != nil {
			if strings.Contains(err.Error(), "RoleAssignmentExists") {
				fmt.Printf("An active or pending role assignment already exists for PIM group '%s'\n", nameFlag)
				os.Exit(0)
			}

			fmt.Printf("Failed to request PIM group activation: %v\n", err)
			os.Exit(1)
		}

		if response.Status.Status != "" {
			fmt.Printf("Success. Status: %s\n", response.Status.Status)
		} else {
			fmt.Printf("Activation request submitted. Response:\n %+v", response)
		}
	},
}

func init() {
	requestCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Name of the PIM group to request activation for (required)")
	requestCmd.Flags().StringVarP(&reasonFlag, "reason", "r", "", "Reason for requesting activation, default is a standard message")
	requestCmd.Flags().DurationVarP(&durationFlag, "duration", "d", 12*time.Hour, "Duration for the activation (e.g., 30m, 1h, 2h)")

	_ = requestCmd.MarkFlagRequired("name")
}
