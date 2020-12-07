package cmd

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	rootCaCsrTemplate = "/etc/containerlab/templates/ca/csr-root-ca.json"
	certCsrTemplate   = "/etc/containerlab/templates/ca/csr.json"
)

var debug bool
var timeout time.Duration
var accessKey string
var secretKey string
var region string
var vpc string

// path to the topology file
var config string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nuage-aws-networkmgr",
	Short: "Automates the connectivity of the aws network manager with nuage",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		envAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		envSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		envRegion := os.Getenv("AWS_DEFAULT_REGION")
		envVPC := os.Getenv("AWS_VPC_NAME")
		log.WithFields(log.Fields{
			"envAccessKey": envAccessKey,
			"envSecretKey": envSecretKey,
			"envRegion":    envRegion,
			"envVPC":       envVPC,
		}).Debug("Environment info")
		if accessKey == "" {
			accessKey = envAccessKey
		}
		if secretKey == "" {
			secretKey = envSecretKey
		}
		if region == "" {
			region = envRegion
		}
		if vpc == "" {
			vpc = envVPC
		}
		if accessKey == "" {
			log.Error("Access Key required e.g. export AWS_ACCESS_KEY_ID='XXXXXXXXXXXX'")
			os.Exit(1)
		}
		if secretKey == "" {
			log.Error("secretKey Key required e.g. export AWS_SECRET_ACCESS_KEY='YYYYYYYYYYYY' ")
			os.Exit(1)
		}
		if region == "" {
			log.Error("region required e.g. export AWS_DEFAULT_REGION='eu-central-1'")
			os.Exit(1)
		}
		if vpc == "" {
			log.Error("vpc required e.g. export AWS_VPC_NAME='eks-nokia-paco-vpc'")
			os.Exit(1)
		}
		if debug {
			log.SetLevel(log.DebugLevel)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "path to the file with configuration information")

}
