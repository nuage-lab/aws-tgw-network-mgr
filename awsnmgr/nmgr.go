package awsnmgr

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/henderiw/nuage-wrapper/pkg/vspk"
	"github.com/nuagenetworks/go-bambou/bambou"

	log "github.com/sirupsen/logrus"
)

// VSD crednetials
//var vsdURL = "https://195.207.5.78:8443"
var vsdUser = "csproot"
var vsdPass = "csproot"
var vsdEnterprise = "csp"

var psk = "AlcatelDC"

// NMgr structure that holds the structure of the network manager
type NMgr struct {
	Config      *Config
	ConfigFile  *string
	Sites       map[string]*Site
	Devices     map[string]*Device
	Connections map[int]*Connection

	Region          *string
	GlobalNetworkID *string

	ClientNMgr *networkmanager.Client
	ClientEC2  map[string]*ec2.Client
	VsdUsr     *vspk.Me

	ctx context.Context

	debug   bool
	timeout time.Duration
}

// Site is a struct that contains the information of a site element
type Site struct {
	Name      string
	SiteID    *string
	Index     int
	Street    string
	Number    int
	City      string
	State     string
	Country   string
	Devices   map[string]*Device
	Endpoints map[string]*Endpoint
}

// Device is a struct that contains the information of a device element
type Device struct {
	Name           string
	DeviceID       *string
	DeviceARN      *string
	NuageNSGateway *vspk.NSGateway
	Index          int
	Kind           string
	Model          string
	Serial         string
	Vendor         string
	Region         string
	Site           *Site
	Endpoints      map[string]*Endpoint
}

// Connection is a struct that contains the information of a link between 2 containers
type Connection struct {
	A      *Endpoint
	B      *Endpoint
	Labels map[string]string
}

// Endpoint is a struct that contains information of a link endpoint
type Endpoint struct {
	Device             *Device
	Site               *Site
	Name               string
	LinkID             *string
	LinkARN            *string
	NuagePort          *vspk.NSPort
	NuageVlan          *vspk.VLAN
	Provider           string
	BwUp               int32
	BwDown             int32
	Kind               string
	PublicIP           string
	Asn                int32
	Cidr               string
	Region             string
	CustomerGatewayID  *string
	CustomerGatewayARN *string
	CustomerGatewayIP  []string
	VPNConnState       string
}

// Option struct
type Option func(nm *NMgr)

// WithDebug function
func WithDebug(d bool) Option {
	return func(n *NMgr) {
		n.debug = d
	}
}

// WithTimeout function
func WithTimeout(dur time.Duration) Option {
	return func(nm *NMgr) {
		nm.timeout = dur
	}
}

// WithConfigFile function
func WithConfigFile(file string) Option {
	return func(nm *NMgr) {
		if file == "" {
			return
		}
		log.Info(file)
		if err := nm.GetTopology(file); err != nil {
			log.Fatalf("failed to read topology file: %v", err)
		}
	}
}

// NewAWsNMgrNuage function defines a new dns-proxy
func NewAWsNMgrNuage(opts ...Option) (*NMgr, error) {
	nm := &NMgr{
		Config:     new(Config),
		ConfigFile: new(string),
		ctx:        context.Background(),
	}
	for _, o := range opts {
		o(nm)
	}

	if nm.Config.Aws.Profile == "" {
		nm.Config.Aws.Profile = "default"
	}

	cfg, err := config.LoadDefaultConfig(
		config.WithRegion("us-west-2"),
		config.WithSharedConfigProfile(nm.Config.Aws.Profile))
	if err != nil {
		panic("failed to load config, " + err.Error())
	}

	nm.Region = &cfg.Region
	nm.ClientNMgr = networkmanager.NewFromConfig(cfg)

	var s *bambou.Session
	s, nm.VsdUsr = vspk.NewSession(vsdUser, vsdPass, vsdEnterprise, nm.Config.Nuage.URL)
	if err := s.Start(); err != nil {
		log.Fatalf("Unable to connect to Nuage VSD: %s", err.Description)
	}

	return nm, nil
}
