package awsnmgr

import (
	"net"

	log "github.com/sirupsen/logrus"

	nuagewrapper "github.com/henderiw/nuage-wrapper"
	"github.com/henderiw/nuage-wrapper/pkg/vspk"
	"github.com/nuagenetworks/go-bambou/bambou"
)

func (nm *NMgr) enterpriseProfile(name string) *vspk.EnterpriseProfile {
	enterpriseProfileCfg := map[string]interface{}{
		"Name":                                   name,
		"Description":                            name,
		"BGPEnabled":                             true,
		"VNFManagementEnabled":                   true,
		"WebFilterEnabled":                       true,
		"AllowAdvancedQOSConfiguration":          true,
		"AllowGatewayManagement":                 true,
		"AllowTrustedForwardingClass":            true,
		"EnableApplicationPerformanceManagement": true,
		"EncryptionManagementMode":               "MANAGED",
		"FloatingIPsQuota":                       1024,
		"DHCPLeaseInterval":                      24,
		"AllowedForwardingClasses":               []interface{}{'A', 'B', 'C', 'D', 'E', 'F', 'G'},
	}

	return nuagewrapper.Enterpriseprofile(enterpriseProfileCfg, nm.VsdUsr)
}

func (nm *NMgr) enterprise(name string, localAS int, entProfile *vspk.EnterpriseProfile) *vspk.Enterprise {
	enterpriseCfg := map[string]interface{}{
		"Name":                  name,
		"LocalAS":               localAS,
		"EnterpriseProfileID":   entProfile.ID,
		"FlowCollectionEnabled": "ENABLED",
	}

	return nuagewrapper.Enterprise(enterpriseCfg, nm.VsdUsr)
}

func (nm *NMgr) getEnterprise(name string) *vspk.Enterprise {
	enterpriseCfg := map[string]interface{}{
		"Name": name,
	}
	return nuagewrapper.GetEnterprise(enterpriseCfg, nm.VsdUsr)
}

func (nm *NMgr) deleteEnterprise(name string) *vspk.Enterprise {
	enterpriseCfg := map[string]interface{}{
		"Name": name,
	}
	return nuagewrapper.DeleteEnterprise(enterpriseCfg, nm.VsdUsr)
}

func (nm *NMgr) domainTemplate(name string, enterprise *vspk.Enterprise) *vspk.DomainTemplate {
	domainTemplateCfg := map[string]interface{}{
		"Name":       name,
		"DPI":        "ENABLED",
		"Encryption": "ENABLED",
	}

	return nuagewrapper.DomainTemplate(domainTemplateCfg, enterprise)
}

func (nm *NMgr) domain(name string, domainTemplate *vspk.DomainTemplate, enterprise *vspk.Enterprise) *vspk.Domain {
	domainCfg := map[string]interface{}{
		"Name":       name,
		"DPI":        "ENABLED",
		"Encryption": "ENABLED",
		"TemplateID": domainTemplate.ID,
	}

	return nuagewrapper.Domain(domainCfg, enterprise)
}

func (nm *NMgr) zone(name string, domain *vspk.Domain) *vspk.Zone {
	zoneCfg := map[string]interface{}{
		"Name": name,
	}

	return nuagewrapper.Zone(zoneCfg, domain)
}

func (nm *NMgr) subnet(name, ip string, zone *vspk.Zone) *vspk.Subnet {
	ipv4Addr, ipv4Net, _ := net.ParseCIDR(ip)
	log.Debugf("ipv4Addr: %s, ipv4Net:%s \n", ipv4Addr, ipv4Net)
	log.Debugf("ipv4Net IP: %s \n", ipv4Net.IP)
	log.Debugf("ipv4Net Mask: %s \n", ipv4Net.Mask)
	log.Debugf("ipv4Net Mask: %T \n", ipv4Net.Mask)

	mask := net.IPMask(net.ParseIP("255.255.255.0").To4()) // If you have the mask as a string
	//mask := net.IPv4Mask(255,255,255,0) // If you have the mask as 4 integer values

	prefixSize, _ := mask.Size()
	log.Debugf("PrefixSize: %d", prefixSize)

	subnetCfg := map[string]interface{}{
		"Name":            name,
		"UnderlayEnabled": "ENABLED",
		"PATEnabled":      "ENABLED",
		"Address":         ipv4Net.IP.String(),
		"Gateway":         ipv4Addr.String(),
		"Netmask":         "255.255.255.0",
		"EVPNEnabled":     true,
		"Advertise":       true,
		"EnableDHCPv4":    true,
	}
	return nuagewrapper.Subnet(subnetCfg, zone)
}

func (nm *NMgr) ingressACLTemplate(name string, priority int, domain *vspk.Domain) *vspk.IngressACLTemplate {
	ingressACLTemplateCfg := map[string]interface{}{
		"Name":              name,
		"Description":       name,
		"Active":            true,
		"DefaultAllowIP":    true,
		"DefaultAllowNonIP": true,
		"AllowAddressSpoof": true,
		"PolicyState":       "LIVE",
		"Priority":          priority,
		"PriorityType":      "NONE",
	}
	return nuagewrapper.IngressACLTemplate(ingressACLTemplateCfg, domain)
}

func (nm *NMgr) egressACLTemplate(name string, priority int, domain *vspk.Domain) *vspk.EgressACLTemplate {
	egressACLTemplateCfg := map[string]interface{}{
		"Name":                           name,
		"Description":                    name,
		"Active":                         true,
		"DefaultAllowIP":                 true,
		"DefaultAllowNonIP":              true,
		"DefaultInstallACLImplicitRules": true,
		"AllowAddressSpoof":              true,
		"PolicyState":                    "LIVE",
		"Priority":                       priority,
		"PriorityType":                   "NONE",
	}
	return nuagewrapper.EgressACLTemplate(egressACLTemplateCfg, domain)
}

func (nm *NMgr) assignAddressrange(subnet *vspk.Subnet, start, end string) {
	addressRanges, _ := subnet.AddressRanges(&bambou.FetchingInfo{})
	if addressRanges == nil {
		addressRange := &vspk.AddressRange{}
		addressRange.DHCPPoolType = "BRIDGE"
		addressRange.IPType = "IPV4"
		addressRange.MaxAddress = end
		addressRange.MinAddress = start
		subnet.CreateAddressRange(addressRange)
	} else {
		log.Debug("Address Ranges already exist")
	}
}

func (nm *NMgr) assignDhcpOptions(subnet *vspk.Subnet, dns1, dns2 string) {
	dhcpOptions, _ := subnet.DHCPOptions(&bambou.FetchingInfo{})
	if dhcpOptions == nil {
		dhcpOption := &vspk.DHCPOption{}
		dhcpOption.ActualType = 6
		dhcpOption.ActualValues = []interface{}{dns1, dns2}
		subnet.CreateDHCPOption(dhcpOption)
	} else {
		log.Debug("DHCP Options already exist")
	}
}

// NSG function
func (nm *NMgr) nsg(name string, enterprise *vspk.Enterprise) *vspk.NSGateway {
	nsgCfg := map[string]interface{}{
		"Name":                  name,
		"TCPMSSEnabled":         true,
		"TCPMaximumSegmentSize": 1330,
		"NetworkAcceleration":   "PERFORMANCE",
	}
	return nuagewrapper.NSG(nsgCfg, enterprise)
}

func (nm *NMgr) getNsg(name string, enterprise *vspk.Enterprise) *vspk.NSGateway {
	nsgCfg := map[string]interface{}{
		"Name": name,
	}
	return nuagewrapper.NSG(nsgCfg, enterprise)
}

func (nm *NMgr) nsgRedundantGwGroup(name string, nsg1, nsg2 *vspk.NSGateway, enterprise *vspk.Enterprise) *vspk.NSRedundantGatewayGroup {
	nsRedundantGwGroupCfg := map[string]interface{}{
		"Name":           name,
		"GatewayPeer1ID": nsg1.ID,
		"GatewayPeer2ID": nsg2.ID,
	}
	return nuagewrapper.NSGRedundantGwGroup(nsRedundantGwGroupCfg, enterprise)
}

func (nm *NMgr) shuntLink(name string, vlan1, vlan2 *vspk.VLAN, nsRedundantGwGroup *vspk.NSRedundantGatewayGroup) *vspk.ShuntLink {
	shuntLinkCfg := map[string]interface{}{
		"Name":        name,
		"VLANPeer1ID": vlan1.ID,
		"VLANPeer2ID": vlan2.ID,
	}
	return nuagewrapper.ShuntLink(shuntLinkCfg, nsRedundantGwGroup)
}

func (nm *NMgr) redundantPort(name string, nsRedundantGwGroup *vspk.NSRedundantGatewayGroup) *vspk.RedundantPort {
	nsRedundantPortCfg := map[string]interface{}{
		"Name":         name,
		"PhysicalName": name,
		"PortType":     "ACCESS",
		"VLANRange":    "0-4094",
	}
	return nuagewrapper.NSGRedundantPort(nsRedundantPortCfg, nsRedundantGwGroup)
}

func (nm *NMgr) nsgNetworkPort(name string, nsg *vspk.NSGateway) *vspk.NSPort {
	nsgPortCfg := map[string]interface{}{
		"Name":            name,
		"PhysicalName":    name,
		"PortType":        "NETWORK",
		"VLANRange":       "0-4094",
		"EnableNATProbes": true,
		"NATTraversal":    "FULL_NAT",
	}
	return nuagewrapper.NSGPort(nsgPortCfg, nsg)
}

func (nm *NMgr) getNetworkPort(name string, nsg *vspk.NSGateway) *vspk.NSPort {
	nsgPortCfg := map[string]interface{}{
		"Name": name,
	}
	return nuagewrapper.GetNSGPort(nsgPortCfg, nsg)
}

func (nm *NMgr) vlan(vlanID int, port *vspk.NSPort) *vspk.VLAN {
	nsgVLANCfg := map[string]interface{}{
		"Value":       vlanID,
		"Description": "test",
	}
	return nuagewrapper.Vlan(nsgVLANCfg, port)
}

func (nm *NMgr) getVlan(vlanID int, port *vspk.NSPort) *vspk.VLAN {
	nsgVLANCfg := map[string]interface{}{
		"Value": vlanID,
	}
	return nuagewrapper.GetVlan(nsgVLANCfg, port)
}

func (nm *NMgr) localNSGRedundantVLAN(vlanID int, port *vspk.RedundantPort) *vspk.VLAN {
	nsgVLANCfg := map[string]interface{}{
		"Value": vlanID,
	}
	return nuagewrapper.RedundantVlan(nsgVLANCfg, port)
}

func (nm *NMgr) staticRoute(domain *vspk.Domain, prefix, nextHop string) {
	ipv4Addr, ipv4Net, _ := net.ParseCIDR(prefix)
	log.Debugf("ipv4Addr: %s\n", ipv4Addr.String())
	log.Debugf("ipv4Net IP: %s \n", ipv4Net.IP.String())
	log.Debugf("ipv4Net Mask: %s \n", ipv4Net.Mask.String())
	log.Debugf("inexthop: %s \n", nextHop)

	staticRouteCfg := map[string]interface{}{
		"Address":   "0.0.0.0",
		"Netmask":   "0.0.0.0",
		"NextHopIp": nextHop,
		"Type":      "OVERLAY",
		"IPType":    "IPV4",
	}
	nuagewrapper.StaticRoute(staticRouteCfg, domain)
}

func (nm *NMgr) bgpNeighbor(subnet *vspk.Subnet, name, peerIP string, peerAS int) {
	bgpNeighborCfg := map[string]interface{}{
		"Name":        name,
		"Description": name,
		"PeerAS":      peerAS,
		"PeerIP":      peerIP,
		"IPType":      "IPV4",
	}
	nuagewrapper.BGPNeighbor(bgpNeighborCfg, subnet)
}

func (nm *NMgr) assignVportBridge(name string, subnet *vspk.Subnet, vlan *vspk.VLAN) *vspk.VPort {
	var vport *vspk.VPort
	vports, _ := subnet.VPorts(&bambou.FetchingInfo{})

	if vports == nil {
		log.Debug("vport does not exist yet")
		vport := &vspk.VPort{}
		vport.Name = name
		vport.VLANID = vlan.ID
		vport.AddressSpoofing = "ENABLED"
		vport.Type = "BRIDGE"
		subnet.CreateVPort(vport)
	} else {
		log.Debug("vport already exist")
		vport = vports[0]
	}
	log.Debugf("vport: %#v \n", vport)
	return vport
}

func (nm *NMgr) assignBridgeInterface(name string, vport *vspk.VPort) {
	log.Debugf("assign bridge interface Name: %s \n", name)
	bridgeInterfaces, _ := vport.BridgeInterfaces(&bambou.FetchingInfo{})

	if bridgeInterfaces == nil {
		bridgeInterface := &vspk.BridgeInterface{}
		bridgeInterface.Name = name
		bridgeInterface.VPortID = vport.ID
		vport.CreateBridgeInterface(bridgeInterface)
	} else {
		log.Debug("bridge Interface already exist")
	}
}

func (nm *NMgr) createIKEPSK(name string, enterprise *vspk.Enterprise) *vspk.IKEPSK {
	ikePSKCfg := map[string]interface{}{
		"Name":           name,
		"Description":    name,
		"UnencryptedPSK": psk,
	}
	return nuagewrapper.CreateIKEPSK(ikePSKCfg, enterprise)
}

func (nm *NMgr) deleteIKEPSK(name string, enterprise *vspk.Enterprise) error {
	ikePSKCfg := map[string]interface{}{
		"Name": name,
	}
	return nuagewrapper.DeleteIKEPSK(ikePSKCfg, enterprise)
}

func (nm *NMgr) createIKEGateway(name, version, ip string, enterprise *vspk.Enterprise) *vspk.IKEGateway {
	ikeGatewayCfg := map[string]interface{}{
		"Name":        name,
		"Description": name,
		"IKEVersion":  version,
		"IPAddress":   ip,
	}
	return nuagewrapper.CreateIKEGateway(ikeGatewayCfg, enterprise)
}

func (nm *NMgr) deleteIKEGateway(name string, enterprise *vspk.Enterprise) error {
	ikeGatewayCfg := map[string]interface{}{
		"Name": name,
	}
	return nuagewrapper.DeleteIKEGateway(ikeGatewayCfg, enterprise)
}

func (nm *NMgr) createIKEEncryptionprofile(name string, enterprise *vspk.Enterprise) *vspk.IKEEncryptionprofile {
	ikeEncryptionProfileCfg := map[string]interface{}{
		"Name":                              name,
		"Description":                       name,
		"DPDMode":                           "REPLY_ONLY",
		"ISAKMPAuthenticationMode":          "PRE_SHARED_KEY",
		"ISAKMPDiffieHelmanGroupIdentifier": "GROUP_2_1024_BIT_DH",
		"ISAKMPEncryptionAlgorithm":         "AES128",
		"ISAKMPEncryptionKeyLifetime":       28800,
		"ISAKMPHashAlgorithm":               "SHA1",
		"IPsecEnablePFS":                    true,
		"IPsecEncryptionAlgorithm":          "AES128",
		"IPsecPreFragment":                  true,
		"IPsecSALifetime":                   3600,
		"IPsecAuthenticationAlgorithm":      "HMAC_SHA1",
		"IPsecSAReplayWindowSize":           "WINDOW_SIZE_64",
	}

	return nuagewrapper.CreateIKEEncryptionProfile(ikeEncryptionProfileCfg, enterprise)
}

func (nm *NMgr) deleteIKEEncryptionprofile(name string, enterprise *vspk.Enterprise) error {
	ikeEncryptionProfileCfg := map[string]interface{}{
		"Name": name,
	}

	return nuagewrapper.DeleteIKEEncryptionProfile(ikeEncryptionProfileCfg, enterprise)
}

func (nm *NMgr) createIKEGatewayProfile(name, pskID, ip, ikeGWID, ikeProfID string, enterprise *vspk.Enterprise) *vspk.IKEGatewayProfile {
	ikeGatewayProfileCfg := map[string]interface{}{
		"Name":                             name,
		"Description":                      name,
		"AssociatedIKEAuthenticationID":    pskID,
		"IKEGatewayIdentifier":             ip,
		"IKEGatewayIdentifierType":         "ID_IPV4_ADDR",
		"AssociatedIKEGatewayID":           ikeGWID,
		"AssociatedIKEEncryptionProfileID": ikeProfID,
	}

	return nuagewrapper.CreateIKEGatewayProfile(ikeGatewayProfileCfg, enterprise)
}

func (nm *NMgr) deleteIKEGatewayProfile(name string, enterprise *vspk.Enterprise) error {
	ikeGatewayProfileCfg := map[string]interface{}{
		"Name": name,
	}

	return nuagewrapper.DeleteIKEGatewayProfile(ikeGatewayProfileCfg, enterprise)
}

func (nm *NMgr) createIKEGatewayConnection(name, id, ikeProfID, pskID string, vlan *vspk.VLAN) *vspk.IKEGatewayConnection {
	ikeGatewayConnCfg1 := map[string]interface{}{
		"Name":                          name,
		"Description":                   name,
		"NSGIdentifier":                 id,
		"NSGIdentifierType":             "ID_KEY_ID",
		"NSGRole":                       "INITIATOR",
		"AllowAnySubnet":                true,
		"AssociatedIKEGatewayProfileID": ikeProfID,
		"AssociatedIKEAuthenticationID": pskID,
	}

	return nuagewrapper.CreateIKEGatewayConnection(ikeGatewayConnCfg1, vlan)
}

func (nm *NMgr) deleteIKEGatewayConnection(name string, vlan *vspk.VLAN) error {
	ikeGatewayConnCfg1 := map[string]interface{}{
		"Name": name,
	}
	return nuagewrapper.DeleteIKEGatewayConnection(ikeGatewayConnCfg1, vlan)
}
