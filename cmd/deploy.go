package cmd

import (
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:          "deploy",
	Short:        "deploy a configuration element",
	Aliases:      []string{"dep"},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
