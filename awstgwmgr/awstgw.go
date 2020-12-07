package awsnmgr

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"

	log "github.com/sirupsen/logrus"
)

// NMgr structure that holds the structure of the network manager
type NMgr struct {
	Config     *Config
	ConfigFile *string

	ClientNMgr *networkmanager.Client

	ctx context.Context

	debug   bool
	timeout time.Duration
}

// Config struct
type Config struct {
	LanInterfaces []string `yaml:"lan-interfaces"`
}

// Option struct
type Option func(n *NMgr)

// WithDebug function
func WithDebug(d bool) Option {
	return func(n *NMgr) {
		n.debug = d
	}
}

// WithTimeout function
func WithTimeout(dur time.Duration) Option {
	return func(n *NMgr) {
		n.timeout = dur
	}
}

// WithConfigFile function
func WithConfigFile(file string) Option {
	return func(n *NMgr) {
		if file == "" {
			return
		}
		log.Info(file)
		//if err := tgw.GetTopology(file); err != nil {
		//	log.Fatalf("failed to read topology file: %v", err)
		//}
	}
}

// WithSecrets function
// func WithSecrets(a, s, r *string) Option {
// 	return func(tgw *Tgw) {
// 		tgw.accessKey = a
// 		tgw.secretKey = s
// 		tgw.region = r
// 	}
// }

// NewAWsNMgrNuage function defines a new dns-proxy
func NewAWsNMgrNuage(opts ...Option) (*NMgr, error) {
	n := &NMgr{
		Config:     new(Config),
		ConfigFile: new(string),
		ctx:        context.Background(),
	}
	for _, o := range opts {
		o(n)
	}

	cfg, err := config.LoadDefaultConfig(
		config.WithRegion("global"))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	n.ClientNMgr = networkmanager.NewFromConfig(cfg)

	return n, nil
}
