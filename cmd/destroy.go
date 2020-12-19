package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:          "destroy",
	Short:        "destroy a configuration element",
	Aliases:      []string{"des"},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("destroying a configuration ...")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
