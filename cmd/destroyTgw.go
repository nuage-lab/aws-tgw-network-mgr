package cmd

import (
	"github.com/nuage-lab/aws-tgw-network-mgr/awsnmgr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// destroyTgwCmd represents the destroy command
var destroyTgwCmd = &cobra.Command{
	Use:          "tgw",
	Short:        "destroy tgw configuration",
	Aliases:      []string{"des"},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("destroying nuage aws tgw network manager configuration ...")
		opts := []awsnmgr.Option{
			awsnmgr.WithDebug(debug),
			awsnmgr.WithTimeout(timeout),
			awsnmgr.WithConfigFile(config),
			//awstgwmgr.WithSecrets(&accessKey, &secretKey, &region),
		}

		nm, err := awsnmgr.NewAWsNMgrNuage(opts...)
		if err != nil {
			log.Fatal(err)
		}

		// Parse topology information
		if err = nm.ParseTopology(); err != nil {
			return err
		}

		// Create AWS resources
		if err := nm.DeleteAWSNetworkMgrNetwork(); err != nil {
			log.Error(err)
		}

		return nil
	},
}

func init() {
	destroyCmd.AddCommand(destroyTgwCmd)
}
