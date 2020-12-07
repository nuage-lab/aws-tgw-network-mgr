package cmd

import (
	"github.com/nuage-lab/aws-tgw-network-mgr/awsnmgr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:          "deploy",
	Short:        "deploy a configuration",
	Aliases:      []string{"dep"},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("deploying nuage aws tgw network managmer configuration ...")
		opts := []awstgwmgr.Option{
			awstgwmgr.WithDebug(debug),
			awstgwmgr.WithTimeout(timeout),
			awstgwmgr.WithConfigFile(config),
			//awstgwmgr.WithSecrets(&accessKey, &secretKey, &region),
		}

		n, err := awsnmgr.NewAWsNMgrNuage(opts...)
		if err != nil {
			log.Fatal(err)
		}

		name := "NuageTestNetwork"
		resp, err := n.CreateGlobalNetwork(&name)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Response: %v", resp)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
