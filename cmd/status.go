package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List both active and pending PIM groups",
	Long:  `List all active and pending PIM groups for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		activeCmd.Run(cmd, args)
		pendingCmd.Run(cmd, args)
	},
}
