package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List both active & pending group activations",
	Long:  `List all active & pending PIM group activations for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		activeCmd.Run(cmd, args)
		pendingCmd.Run(cmd, args)
	},
}
