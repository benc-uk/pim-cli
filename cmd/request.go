package cmd

import (
	"context"
	"log"
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
			reasonFlag = "Standard activation request via pimg-cli for " + userName
		}

		err = pim.RequestPIMGroupActivation(ctx, cred, userID, nameFlag, reasonFlag, durationFlag)
		if err != nil {
			log.Fatalf("Failed to request PIM group activation: %v", err)
		}
	},
}

func init() {
	requestCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Name of the PIM group to request activation for (required)")
	requestCmd.Flags().StringVarP(&reasonFlag, "reason", "r", "", "Reason for requesting activation, default is a standard message")
	requestCmd.Flags().DurationVarP(&durationFlag, "duration", "d", 12*time.Hour, "Duration for the activation (e.g., 30m, 1h, 2h)")

	requestCmd.MarkFlagRequired("name")
}
