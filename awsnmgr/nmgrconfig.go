package awsnmgr

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// supported kinds
var kinds = []string{"sdwan", "tgw"}

// Config defines lab configuration as it is provided in the YAML file
type Config struct {
	Name     string   `json:"name,omitempty"`
	Nuage    Nuage    `json:"nuage,omitempty"`
	Aws      Aws      `json:"aws,omitempty"`
	Topology Topology `json:"topology,omitempty"`
}

// Aws related information
type Aws struct {
	Profile string `json:"profile,omitempty"`
}

// Nuage related information
type Nuage struct {
	Enterprise string `json:"enterprise,omitempty"`
	URL        string `json:"url,omitempty"`
}

// Topology represents a lab topology
type Topology struct {
	Sites       map[string]SiteConfig   `yaml:"sites,omitempty"`
	DeviceKinds map[string]DeviceConfig `yaml:"device-kinds,omitempty"`
	Devices     map[string]DeviceConfig `yaml:"devices,omitempty"`
	Connections []ConnectionConfig      `yaml:"connections,omitempty"`
}

// DeviceConfig represents a configuration a given device can have
type DeviceConfig struct {
	Kind   string `yaml:"kind,omitempty"`
	Vendor string `yaml:"vendor,omitempty"`
	Model  string `yaml:"model,omitempty"`
	Serial string `yaml:"serial,omitempty"`
	Region string `yaml:"region,omitempty"`
}

// ConnectionConfig struct
type ConnectionConfig struct {
	Endpoints []string
	Labels    map[string]string `yaml:"labels,omitempty"`
}

// SiteConfig represents a configuration a given site can have
type SiteConfig struct {
	Street  string `yaml:"street,omitempty"`
	Number  int    `yaml:"number,omitempty"`
	City    string `yaml:"city,omitempty"`
	State   string `yaml:"state,omitempty"`
	Country string `yaml:"country,omitempty"`
}

// GetTopology parses the topology file into c.Conf structure
// as well as populates the TopoFile structure with the topology file related information
func (nm *NMgr) GetTopology(topo string) error {
	log.Infof("Getting topology information from %s file...", topo)

	yamlFile, err := ioutil.ReadFile(topo)
	if err != nil {
		return err
	}
	log.Debugf(fmt.Sprintf("Topology file contents:\n%s\n", yamlFile))

	err = yaml.Unmarshal(yamlFile, nm.Config)
	if err != nil {
		return err
	}

	return nil
}

// ParseTopology parses the configuration topology
func (nm *NMgr) ParseTopology() error {
	log.Info("Parsing topology information ...")
	log.Debugf("Lab name: %s", nm.Config.Name)

	// initialize Sites, Devices and Connection variable
	nm.Sites = make(map[string]*Site)
	nm.Devices = make(map[string]*Device)
	nm.Connections = make(map[int]*Connection)
	nm.ClientEC2 = make(map[string]*ec2.Client)

	// initialize the Site information from the topology file
	idx := 0
	for name, site := range nm.Config.Topology.Sites {
		log.Debugf("Site info: %d, %s, %v", idx, name, site)

		if err := nm.NewSite(name, site, idx); err != nil {
			return err
		}
		idx++
	}
	// initialize the Device information from the topology file
	idx = 0
	for name, device := range nm.Config.Topology.Devices {
		log.Debugf("Device info: %d, %s, %s", idx, name, device)

		if err := nm.NewDevice(name, device, idx); err != nil {
			return err
		}
		idx++
	}
	for i, c := range nm.Config.Topology.Connections {
		log.Debugf("Connection info: %d, %v", i, c)
		// i represents the endpoint integer and c provide the connection struct
		nm.Connections[i] = nm.NewConnection(c)
	}
	return nil
}

// NewSite initializes a new site object
func (nm *NMgr) NewSite(name string, cfg SiteConfig, idx int) error {
	// initialize a new node
	s := new(Site)
	s.Name = name
	s.Index = idx
	s.Street = cfg.Street
	s.Number = cfg.Number
	s.City = cfg.City
	s.State = cfg.State
	s.Country = cfg.Country

	s.Devices = make(map[string]*Device)
	s.Endpoints = make(map[string]*Endpoint)

	nm.Sites[name] = s
	return nil
}

// NewDevice initializes a new device object
func (nm *NMgr) NewDevice(name string, cfg DeviceConfig, idx int) error {
	// initialize a new node
	d := new(Device)

	d.Name = name
	d.Kind = cfg.Kind

	switch d.Kind {
	case "sdwan":
		d.Vendor = cfg.Vendor
		d.Serial = cfg.Serial
		d.Model = cfg.Model
	case "tgw":
		d.Region = cfg.Region
		cfg, err := config.LoadDefaultConfig(
			config.WithRegion(cfg.Region),
			config.WithSharedConfigProfile("admin"))
		if err != nil {
			panic("failed to load config, " + err.Error())
		}
		nm.ClientEC2[cfg.Region] = ec2.NewFromConfig(cfg)

	default:
		return fmt.Errorf("Node '%s' refers to a kind '%s' which is not supported. Supported kinds are %q", name, d.Kind, kinds)
	}

	d.Site = new(Site)
	d.Endpoints = make(map[string]*Endpoint)

	nm.Devices[name] = d
	return nil
}

// NewConnection initializes a new link object
func (nm *NMgr) NewConnection(cCfg ConnectionConfig) *Connection {
	// initialize a new link
	c := new(Connection)
	c.Labels = cCfg.Labels

	for i, d := range cCfg.Endpoints {
		// i indicates the number and d presents the string, which need to be
		// split in node and endpoint name
		if i == 0 {
			c.A = nm.NewEndpoint(i, d, c.Labels)
		} else {
			c.B = nm.NewEndpoint(i, d, c.Labels)
		}
	}
	// map the region from link B to link A
	if c.B.Device.Kind == "tgw" {
		c.A.Region = c.B.Region
	}
	return c
}

// NewEndpoint initializes a new endpoint object
func (nm *NMgr) NewEndpoint(i int, e string, l map[string]string) *Endpoint {
	// initialize a new endpoint
	endpoint := new(Endpoint)

	siteName := ""
	deviceName := ""
	epName := ""
	// split the string to get node name and endpoint name
	split := strings.Split(e, ":")
	switch len(split) {
	case 3: // sdwan site
		siteName = split[0]   // site name
		deviceName = split[1] // device name
		epName = split[2]     // endpoint name
		if _, ok := l["provider"]; ok {
			endpoint.Provider = l["provider"]
		}
		if _, ok := l["bwdown"]; ok {
			bwDown, err := strconv.Atoi(l["bwdown"])
			if err != nil {
				log.Errorf("strconv fails: %s", err)
			}
			endpoint.BwDown = int32(bwDown)
		}
		if _, ok := l["bwup"]; ok {
			bwUp, err := strconv.Atoi(l["bwup"])
			if err != nil {
				log.Errorf("strconv fails: %s", err)
			}
			endpoint.BwUp = int32(bwUp)
		}
		if _, ok := l["kind"]; ok {
			endpoint.Kind = l["kind"]
		}
		if _, ok := l["public-ip"]; ok {
			endpoint.PublicIP = l["public-ip"]
		}
		if _, ok := l["asn"]; ok {
			asn, err := strconv.Atoi(l["asn"])
			if err != nil {
				log.Errorf("strconv fails: %s", err)
			}
			endpoint.Asn = int32(asn)
		}
		if _, ok := l["cidr"]; ok {
			endpoint.Cidr = l["cidr"]
		}
	case 1: // transit gateway
		if i == 1 {
			deviceName = e
			epName = e
		} else {
			log.Fatalf("endpoint %s has wrong syntax", e)
		}
	default:
		log.Fatalf("endpoint %s has wrong syntax", e)
	}

	for name, s := range nm.Sites {
		if name == siteName {
			endpoint.Site = s
			break
		}
	}
	// search the node device based in the name of the split function
	for name, d := range nm.Devices {
		if name == deviceName {
			endpoint.Device = d
			endpoint.Region = d.Region
			endpoint.Name = siteName + "-" + deviceName + "-" + epName
			break
		}
	}

	if endpoint.Device == nil {
		log.Fatalf("Not all nodes are specified in the 'topology.nodes' section or the names don't match in the 'links.endpoints' section: %s", deviceName)
	}
	log.Debugf("Endpoints Info: %s, %s, %s", siteName, deviceName, epName)

	if endpoint.Site != nil {
		endpoint.Device.Site = endpoint.Site
		endpoint.Site.Devices[deviceName] = endpoint.Device
		endpoint.Site.Endpoints[epName] = endpoint
	}
	endpoint.Device.Endpoints[epName] = endpoint

	return endpoint
}

// CreateAWSNetworkMgrNetwork function
func (nm *NMgr) CreateAWSNetworkMgrNetwork() error {
	log.Infof("Create Global Network: %s", nm.Config.Name)
	respNetw, err := nm.CreateGlobalNetwork(&nm.Config.Name)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Global Network Id: %v", *respNetw.GlobalNetwork.GlobalNetworkId)
	nm.GlobalNetworkID = respNetw.GlobalNetwork.GlobalNetworkId

	for deviceName, device := range nm.Devices {
		switch device.Kind {
		case "tgw":
			log.Infof("Create TGW: %s", deviceName)
			r, err := nm.CreateTransitGateway(&device.Region, &deviceName)
			if err != nil {
				log.Fatalf("Error create device: %s", err)
			}
			log.Infof("Device Id: %v", *r.TransitGateway.TransitGatewayId)
			device.DeviceID = r.TransitGateway.TransitGatewayId
			device.DeviceARN = r.TransitGateway.TransitGatewayArn
			_, err = nm.RegisterTransitGateway(device.DeviceARN)
			if err != nil {
				log.Errorf("Error associating TGW: %s", err)
			}

		}
	}
	return nil
}

// DeleteAWSNetworkMgrNetwork function
func (nm *NMgr) DeleteAWSNetworkMgrNetwork() error {
	r, err := nm.DescribeGlobalNetworks()
	if err != nil {
		log.Fatal(err)
	}

	//if len(r.GlobalNetworks) > 0 {
	for idx, g := range r.GlobalNetworks {
		for i := 0; i < len(g.Tags); i++ {
			if *g.Tags[i].Key == "Name" {
				if *g.Tags[i].Value == nm.Config.Name {
					log.Infof("Global Network found")
					nm.GlobalNetworkID = r.GlobalNetworks[idx].GlobalNetworkId
				}
			}
		}
	}
	if nm.GlobalNetworkID != nil {
		for deviceName, device := range nm.Devices {
			switch device.Kind {
			case "tgw":
				r, err := nm.GetTransitGatewayRegistrations()
				if err != nil {
					log.Error(err)
				}
				for _, t := range r.TransitGatewayRegistrations {
					log.Infof("Deregister Transit Gateway....")
					_, err = nm.DeregisterTransitGateway(t.TransitGatewayArn)
					if err != nil {
						log.Error(err)
					}
				}
				log.Infof("Delete Transit Gateway....")
				tgws, err := nm.DescribeTransitGateways(&device.Region, &deviceName)
				if err != nil {
					log.Error(err)
				}
				for _, t := range tgws.TransitGateways {
					_, err = nm.DeleteTransitGateway(&device.Region, t.TransitGatewayId)
					if err != nil {
						log.Fatal(err)
					}
				}	
			}
		}
		time.Sleep(30 * time.Second)

		log.Infof("Delete Global Network....")
		if _, err := nm.DeleteGlobalNetwork(); err != nil {
			log.Errorf("Error deleting Global Network: %s", err)
		}
	} else {
		log.Infof("Nothing to delete....")
	}

	return nil
}

// CreateAWSNetworkMgrSites function
func (nm *NMgr) CreateAWSNetworkMgrSites() error {
	log.Infof("Add sites to Global Network: %s", nm.Config.Name)
	respNetw, err := nm.CreateGlobalNetwork(&nm.Config.Name)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Global Network Id: %v", *respNetw.GlobalNetwork.GlobalNetworkId)
	nm.GlobalNetworkID = respNetw.GlobalNetwork.GlobalNetworkId

	enterprise := nm.getEnterprise(nm.Config.Nuage.Enterprise)
	if enterprise == nil {
		log.Errorf("Enterprise does not exist : %s", nm.Config.Nuage.Enterprise)
	}
	log.Debugf("Enterprise ID : %v", enterprise.ID)

	ikePSK := nm.createIKEPSK("AWS-"+nm.Config.Name+"PSK", enterprise)
	log.Debugf("ikePSK: %v", ikePSK.ID)

	ikeEncryptionProfile := nm.createIKEEncryptionprofile("AWS-"+nm.Config.Name, enterprise)
	log.Debugf("ikeEncryptionProfile: %v", ikeEncryptionProfile)

	for siteName, site := range nm.Sites {
		log.Infof("Create Site: %s", siteName)
		r, err := nm.CreateSite(&siteName, site)
		if err != nil {
			log.Fatalf("Error create site: %s", err)
		}
		log.Debugf("Site Id: %v", *r.Site.SiteId)
		site.SiteID = r.Site.SiteId
	}

	for deviceName, device := range nm.Devices {
		switch device.Kind {
		case "sdwan":
			log.Infof("Create Device: %s", deviceName)

			nsGateway := nm.getNsg(deviceName, enterprise)
			if nsGateway == nil {
				log.Errorf("Nuage NSG device does not exist: %s", deviceName)
			}
			log.Debugf("Nuage NSG ID: %s", nsGateway.ID)
			device.NuageNSGateway = nsGateway

			r, err := nm.CreateDevice(&deviceName, device)
			if err != nil {
				log.Fatalf("Error create device: %s", err)
			}
			log.Debugf("Device Id: %v", *r.Device.DeviceId)
			device.DeviceID = r.Device.DeviceId
			device.DeviceARN = r.Device.DeviceArn
			for epName, ep := range device.Endpoints {

				nsgPort := nm.getNetworkPort(epName, nsGateway)
				if nsgPort == nil {
					log.Errorf("Nuage NSG port does not exist: %s", epName)
				}
				log.Debugf("Nuage PORT: %s", nsgPort.ID)
				ep.NuagePort = nsgPort

				nsVlan := nm.getVlan(0, nsgPort)
				if nsVlan == nil {
					log.Errorf("Nuage NSG vlan does not exist: 0")
				}
				log.Debugf("Nuage VLAN: %s", nsVlan.ID)
				ep.NuageVlan = nsVlan

				r, err := nm.CreateLink(&epName, ep)
				if err != nil {
					log.Fatalf("Error create link: %s", err)
				}
				log.Debugf("Link Id: %v", *r.Link.LinkId)
				ep.LinkID = r.Link.LinkId
				_, err = nm.AssociateLink(device.DeviceID, ep.LinkID)
				if err != nil {
					log.Fatalf("Error create link: %s", err)
				}
			}
		case "tgw":
			log.Infof("Find TGW: %s", deviceName)

			state := false
			found := false
			i := 1
			for ok := true; ok; ok = !state {
				r, err := nm.DescribeTransitGateways(&device.Region, &deviceName)
				if err != nil {
					log.Fatalf("Error create device: %s", err)
				}
				if len(r.TransitGateways) == 0 {
					log.Errorf("No transit GWs found, first 'awsnuagenetwmgr run deploy tgw -c <confi-file'")
					return nil
				}

				for i, t := range r.TransitGateways {
					if t.State == "deleted" || t.State == "deleting" {
						// do nothing
					} else {
						found = true
						if t.State == "available" {
							state = true
						}
						device.DeviceID = r.TransitGateways[i].TransitGatewayId
						device.DeviceARN = r.TransitGateways[i].TransitGatewayArn
						log.Debugf("Transit GW Id: %s", *r.TransitGateways[i].TransitGatewayId)
					}
				}
				if !found {
					log.Errorf("No transit GWs found, first 'awsnuagenetwmgr run deploy tgw -c <confi-file'")
					return nil
				}
				if !state {
					log.Infof("Wait a minute to check if all tgw are in available state, total (%d min)", i)
					time.Sleep(60 * time.Second)
					i++
				}

			}

		}
	}

	for _, conn := range nm.Connections {
		if conn.A.Device.Kind == "sdwan" {
			if conn.A.PublicIP != "" {
				log.Infof("Create Customer Gateway: %s %s %s", conn.A.Region, conn.A.Name, conn.A.PublicIP)
				r, err := nm.CreateCustomerGateway(&conn.A.Region, &conn.A.Name, &conn.A.PublicIP, &conn.A.Asn)
				if err != nil {
					log.Fatalf("Error create customer gateway: %s", err)
				}

				if conn.B.Device.Kind == "tgw" {
					log.Infof("Create VPN connection: %s %s %s", conn.A.Region, conn.A.Name, conn.A.Cidr)
					r, err := nm.CreateVpnConnection(&conn.A.Region, &conn.A.Name, r.CustomerGateway.CustomerGatewayId, conn.B.Device.DeviceID, &conn.A.Cidr)
					if err != nil {
						log.Fatalf("Error create vpn connection: %s", err)
					}
					//log.Infof("VPN Connection: %v", *r.VpnConnection.CustomerGatewayConfiguration)
					vpnConn := VpnConnection{}
					xml.Unmarshal([]byte(*r.VpnConnection.CustomerGatewayConfiguration), &vpnConn)
					for i, ipsec := range vpnConn.IpsecTunnel {
						log.Debugf("VPN IP address : %s", ipsec.VpnGateway.TunnelOutsideAddress.IPAddress)
						conn.A.CustomerGatewayIP = append(conn.A.CustomerGatewayIP, ipsec.VpnGateway.TunnelOutsideAddress.IPAddress)

						ikeGatewayCfg := nm.createIKEGateway("TGWCGW"+conn.A.Region+conn.A.Device.Name+conn.A.Name+strconv.Itoa(i), "V1", ipsec.VpnGateway.TunnelOutsideAddress.IPAddress, enterprise)
						log.Debugf("ikeGatewayCfg: %v", ikeGatewayCfg)

						ikeGatewayProfile := nm.createIKEGatewayProfile("TGWCGW"+conn.A.Region+conn.A.Device.Name+conn.A.Name+strconv.Itoa(i), ikePSK.ID, ipsec.VpnGateway.TunnelOutsideAddress.IPAddress, ikeGatewayCfg.ID, ikeEncryptionProfile.ID, enterprise)
						log.Debugf("ikeGatewayProfile: %v", ikeGatewayProfile)

						ikeGatewayconn := nm.createIKEGatewayConnection("TGWCGW"+conn.A.Region+conn.A.Device.Name+conn.A.Name+strconv.Itoa(i), conn.A.Device.Name, ikeGatewayProfile.ID, ikePSK.ID, conn.A.NuageVlan)
						log.Debugf("ikeGatewayconn: %v", ikeGatewayconn)
					}
				}
				log.Debugf("Customer Gateway Id: %v", *r.CustomerGateway.CustomerGatewayId)
				CustomerGatewayArn := "arn:aws:ec2:" + conn.A.Region + ":610303483713:customer-gateway/" + *r.CustomerGateway.CustomerGatewayId
				conn.A.CustomerGatewayID = r.CustomerGateway.CustomerGatewayId
				conn.A.VPNConnState = "not available"
				conn.A.CustomerGatewayARN = &CustomerGatewayArn
				log.Debugf("Customer Gateway ARN: %v", CustomerGatewayArn)

			}
		}
	}

	log.Infof("Checking VPN connection status before we can associate the device/links with the customer GW")
	state := false
	i := 0
	for ok := true; ok; ok = !state {
		i++
		log.Infof("Wait a minute to check if all gw are in available state, total (%d min)", i)
		time.Sleep(60 * time.Second)
		for _, conn := range nm.Connections {
			if conn.A.Device.Kind == "sdwan" {
				if conn.A.PublicIP != "" {
					r, err := nm.DescribeVpnConnections(&conn.A.Region, &conn.A.Name)
					if err != nil {
						log.Errorf("Error get vpn connection: %s", err)
					}
					for _, v := range r.VpnConnections {
						log.Infof("Customer GW state: %s %s %t", conn.A.Name, v.State, state)
						if v.State == types.VpnStateAvailable {
							if conn.A.VPNConnState != "available" {
								// only assocate the connection to the customer gw once
							}
							conn.A.VPNConnState = "available"
						} else {
							conn.A.VPNConnState = "not available"
						}

					}
				}
			}
		}
		// check if all vpn connections are in available status
		for _, conn := range nm.Connections {
			if conn.A.Device.Kind == "sdwan" {
				if conn.A.VPNConnState == "available" {
					state = true
				} else {
					state = false
					break
				}
			}
		}
	}

	time.Sleep(60 * time.Second)

	for _, conn := range nm.Connections {
		if conn.A.Device.Kind == "sdwan" {
			log.Infof("Associate Customer GW: %s %s %s %s", *conn.A.CustomerGatewayARN, *conn.A.Device.DeviceID, *conn.A.LinkID, conn.A.Device.Name)
			_, err = nm.AssociateCustomerGateway(conn.A.CustomerGatewayARN, conn.A.Device.DeviceID, conn.A.LinkID)
			if err != nil {
				log.Fatalf("Error associate customer gateway: %s", err)
			}
		}
	}

	return nil
}

// DeleteAWSNetworkMgrSites function
func (nm *NMgr) DeleteAWSNetworkMgrSites() error {
	r, err := nm.DescribeGlobalNetworks()
	if err != nil {
		log.Fatal(err)
	}

	enterprise := nm.getEnterprise(nm.Config.Nuage.Enterprise)
	if enterprise == nil {
		log.Errorf("Enterprise does not exist : %s", nm.Config.Nuage.Enterprise)
	}
	log.Debugf("Enterprise ID : %v", enterprise.ID)

	//if len(r.GlobalNetworks) > 0 {
	for idx, g := range r.GlobalNetworks {
		for i := 0; i < len(g.Tags); i++ {
			if *g.Tags[i].Key == "Name" {
				if *g.Tags[i].Value == nm.Config.Name {
					log.Debugf("Global Network found")
					nm.GlobalNetworkID = r.GlobalNetworks[idx].GlobalNetworkId
				}
			}
		}
	}
	if nm.GlobalNetworkID != nil {
		// get the site ID, device ID(s), Link ID(s) from the AWS to remove the associations
		r, err := nm.GetSites()
		if err != nil {
			log.Fatal(err)
		}
		for _, sa := range r.Sites {
			for i := 0; i < len(sa.Tags); i++ {
				if *sa.Tags[i].Key == "Name" {
					for _, s := range nm.Sites {
						if *sa.Tags[i].Value == s.Name {
							log.Infof("Site exists")
							s.SiteID = sa.SiteId
							r, err := nm.GetDevice(s.SiteID)
							if err != nil {
								log.Fatal(err)
							}
							l, err := nm.GetLink(s.SiteID)
							if err != nil {
								log.Fatal(err)
							}
							for _, da := range r.Devices {
								for i := 0; i < len(da.Tags); i++ {
									for _, d := range nm.Devices {
										if *da.Tags[i].Value == d.Name {
											log.Debugf("Device exists")
											d.Site = s
											d.DeviceID = da.DeviceId
											d.DeviceARN = da.DeviceArn
										}
										for _, la := range l.Links {
											log.Debugf("AWS LINK INFO: %s, %s", *la.LinkId, *la.Description)
											for i := 0; i < len(la.Tags); i++ {
												for n, ep := range d.Endpoints {
													if *la.Tags[i].Value == n {
														log.Debugf("Link exists on device")
														ep.LinkID = la.LinkId
														ep.LinkARN = la.LinkArn

													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		rc, err := nm.GetCustomerGatewayAssociations()
		for _, c := range rc.CustomerGatewayAssociations {
			log.Infof("Customer gateway DisAssociation: %s, %s, %s", *c.CustomerGatewayArn, *c.DeviceId, *c.LinkId)
			_, err := nm.DisassociateCustomerGateway(c.CustomerGatewayArn, c.DeviceId, c.LinkId)
			if err != nil {
				log.Error(err)
			}
		}
		log.Infof("Disassociating  Links....")
		for deviceName, d := range nm.Devices {
			if d.DeviceID != nil && d.Site.SiteID != nil {
				log.Debugf("Device Name: %s, %s, %s", d.Name, *d.DeviceID, *d.Site.SiteID)

				nsGateway := nm.getNsg(deviceName, enterprise)
				if nsGateway == nil {
					log.Errorf("Nuage NSG device does not exist: %s", deviceName)
				}
				log.Debugf("Nuage NSG ID: %s", nsGateway.ID)
				d.NuageNSGateway = nsGateway

				for epName, ep := range d.Endpoints {
					if ep.LinkID != nil {
						log.Debugf("Link Name: %s, %s", ep.Name, *ep.LinkID)

						nsgPort := nm.getNetworkPort(epName, nsGateway)
						if nsgPort == nil {
							log.Errorf("Nuage NSG port does not exist: %s", epName)
						}
						log.Debugf("Nuage PORT: %s", nsgPort.ID)
						ep.NuagePort = nsgPort

						nsVlan := nm.getVlan(0, nsgPort)
						if nsVlan == nil {
							log.Errorf("Nuage NSG vlan does not exist: 0")
						}
						log.Debugf("Nuage VLAN: %s", nsVlan.ID)
						ep.NuageVlan = nsVlan

						if _, err := nm.DisassociateLink(d.DeviceID, ep.LinkID); err != nil {
							log.Errorf("Error disassociating links: %s", err)
						}
					}
				}
			} else {
				log.Debugf("Device Name: %s", d.Name)
			}
		}

		log.Infof("Deleting Links....")
		if err := nm.DeleteLinks(); err != nil {
			log.Errorf("Error deleting links: %s", err)
		}
		log.Infof("Deleting Devices....")
		if err := nm.DeleteDevices(); err != nil {
			log.Errorf("Error deleting Devices: %s", err)
		}
		log.Infof("Deleting Sites....")
		if err := nm.DeleteSites(); err != nil {
			log.Errorf("Error deleting Sites: %s", err)
		}
		for _, conn := range nm.Connections {
			if conn.A.Device.Kind == "sdwan" {
				if conn.A.PublicIP != "" {
					if conn.B.Device.Kind == "tgw" {
						r, err := nm.DescribeVpnConnections(&conn.A.Region, &conn.A.Name)
						if err != nil {
							log.Fatalf("Error describe vpn connection: %s", err)
						}
						for _, c := range r.VpnConnections {
							log.Infof("Delete Vpn Connection....")
							_, err = nm.DeleteVpnConnection(&conn.A.Region, c.VpnConnectionId)
							if err != nil {
								log.Fatal(err)
							}

							for i := 0; i < 2; i++ {
								err = nm.deleteIKEGatewayConnection("TGWCGW"+conn.A.Region+conn.A.Device.Name+conn.A.Name+strconv.Itoa(i), conn.A.NuageVlan)
								if err != nil {
									log.Errorf("delete deleteIKEGatewayConnection error: %v", err)
								}

								err = nm.deleteIKEGatewayProfile("TGWCGW"+conn.A.Region+conn.A.Device.Name+conn.A.Name+strconv.Itoa(i), enterprise)
								if err != nil {
									log.Errorf("delete ikeGatewayProfile error: %v", err)
								}

								err = nm.deleteIKEGateway("TGWCGW"+conn.A.Region+conn.A.Device.Name+conn.A.Name+strconv.Itoa(i), enterprise)
								if err != nil {
									log.Errorf("deleteIKEGateway error: %v", err)
								}
							}
						}
					}
					r, err := nm.DescribeCustomerGateways(&conn.A.Region, &conn.A.Name)
					if err != nil {
						log.Fatalf("Error describe customer gateway: %s", err)
					}
					for _, c := range r.CustomerGateways {
						log.Infof("Delete Customer Gateway....")
						_, err = nm.DeleteCustomerGateway(&conn.A.Region, c.CustomerGatewayId)
						if err != nil {
							log.Fatal(err)
						}
					}
				}
			}
		}

		err = nm.deleteIKEEncryptionprofile("AWS-"+nm.Config.Name, enterprise)
		if err != nil {
			log.Errorf("Error deleting Encryption profile: %s", err)
		}
		err = nm.deleteIKEPSK("AWS-"+nm.Config.Name+"PSK", enterprise)
		if err != nil {
			log.Errorf("Error deleting PSK: %s", err)
		}

	} else {
		log.Infof("Nothing to delete....")
	}
	return nil
}
