package cmd

import (
	"github.com/nuage-lab/aws-tgw-network-mgr/awsnmgr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deploySitesCmd represents the deploy command
var deploySitesCmd = &cobra.Command{
	Use:          "sites",
	Short:        "deploy site configuration",
	Aliases:      []string{"dep"},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("deploying nuage aws tgw network manager sites configuration ...")
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
		if err := nm.CreateAWSNetworkMgrSites(); err != nil {
			log.Error(err)
		}

		return nil
	},
}

func init() {
	deployCmd.AddCommand(deploySitesCmd)
}
