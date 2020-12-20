package awsnmgr

import (
	"encoding/xml"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
)

// VpnConnection struct
type VpnConnection struct {
	XMLName                 xml.Name `xml:"vpn_connection"`
	Text                    string   `xml:",chardata"`
	ID                      string   `xml:"id,attr"`
	CustomerGatewayID       string   `xml:"customer_gateway_id"`
	VpnGatewayID            string   `xml:"vpn_gateway_id"`
	VpnConnectionType       string   `xml:"vpn_connection_type"`
	VpnConnectionAttributes string   `xml:"vpn_connection_attributes"`
	IpsecTunnel             []struct {
		Text            string `xml:",chardata"`
		CustomerGateway struct {
			Text                 string `xml:",chardata"`
			TunnelOutsideAddress struct {
				Text      string `xml:",chardata"`
				IPAddress string `xml:"ip_address"`
			} `xml:"tunnel_outside_address"`
			TunnelInsideAddress struct {
				Text        string `xml:",chardata"`
				IPAddress   string `xml:"ip_address"`
				NetworkMask string `xml:"network_mask"`
				NetworkCidr string `xml:"network_cidr"`
			} `xml:"tunnel_inside_address"`
		} `xml:"customer_gateway"`
		VpnGateway struct {
			Text                 string `xml:",chardata"`
			TunnelOutsideAddress struct {
				Text      string `xml:",chardata"`
				IPAddress string `xml:"ip_address"`
			} `xml:"tunnel_outside_address"`
			TunnelInsideAddress struct {
				Text        string `xml:",chardata"`
				IPAddress   string `xml:"ip_address"`
				NetworkMask string `xml:"network_mask"`
				NetworkCidr string `xml:"network_cidr"`
			} `xml:"tunnel_inside_address"`
		} `xml:"vpn_gateway"`
		Ike struct {
			Text                   string `xml:",chardata"`
			AuthenticationProtocol string `xml:"authentication_protocol"`
			EncryptionProtocol     string `xml:"encryption_protocol"`
			Lifetime               string `xml:"lifetime"`
			PerfectForwardSecrecy  string `xml:"perfect_forward_secrecy"`
			Mode                   string `xml:"mode"`
			PreSharedKey           string `xml:"pre_shared_key"`
		} `xml:"ike"`
		Ipsec struct {
			Text                          string `xml:",chardata"`
			Protocol                      string `xml:"protocol"`
			AuthenticationProtocol        string `xml:"authentication_protocol"`
			EncryptionProtocol            string `xml:"encryption_protocol"`
			Lifetime                      string `xml:"lifetime"`
			PerfectForwardSecrecy         string `xml:"perfect_forward_secrecy"`
			Mode                          string `xml:"mode"`
			ClearDfBit                    string `xml:"clear_df_bit"`
			FragmentationBeforeEncryption string `xml:"fragmentation_before_encryption"`
			TCPMssAdjustment              string `xml:"tcp_mss_adjustment"`
			DeadPeerDetection             struct {
				Text     string `xml:",chardata"`
				Interval string `xml:"interval"`
				Retries  string `xml:"retries"`
			} `xml:"dead_peer_detection"`
		} `xml:"ipsec"`
	} `xml:"ipsec_tunnel"`
}

func createEC2Tags(tagKey, tagValue *string) (tags []types.Tag) {
	tag := types.Tag{
		Key:   tagKey,
		Value: tagValue,
	}
	tags = append(tags, tag)
	return tags
}

func createEC2TagSpecs(tagKey, tagValue *string, rt types.ResourceType) (tspecs []types.TagSpecification) {
	t := createEC2Tags(tagKey, tagValue)

	tspec := types.TagSpecification{
		ResourceType: rt,
		Tags:         t,
	}
	tspecs = append(tspecs, tspec)
	return tspecs
}

func createEC2Filter(tagKey, tagValue *string) (filters []types.Filter) {
	var values []string
	values = append(values, *tagValue)

	filter := types.Filter{
		Name:   tagKey,
		Values: values,
	}
	filters = append(filters, filter)
	return filters

}

// CreateTransitGateway fucntion
func (nm *NMgr) CreateTransitGateway(region, name *string) (*ec2.CreateTransitGatewayOutput, error) {
	
	r, err := nm.DescribeTransitGateways(region, name)
	if err != nil {
		log.Fatal(err)
	}

	for i, t := range r.TransitGateways {
		if t.State == "deleted" || t.State == "deleting" {
			
		} else {
			log.Infof("Transit Gateway exists")
			o := &ec2.CreateTransitGatewayOutput{
				TransitGateway: &r.TransitGateways[i],
			}
			return o, nil
		}
	}


	tagKey := "Name"
	tspecs := createEC2TagSpecs(&tagKey, name, types.ResourceTypeTransitGateway)

	o := &types.TransitGatewayRequestOptions{
		AmazonSideAsn:                64512,
		AutoAcceptSharedAttachments:  types.AutoAcceptSharedAttachmentsValueDisable,
		DefaultRouteTableAssociation: types.DefaultRouteTableAssociationValueEnable,
		DefaultRouteTablePropagation: types.DefaultRouteTablePropagationValueEnable,
		DnsSupport:                   types.DnsSupportValueEnable,
		MulticastSupport:             types.MulticastSupportValueDisable,
		VpnEcmpSupport:               types.VpnEcmpSupportValueEnable,
	}

	input := &ec2.CreateTransitGatewayInput{
		Description:       name,
		Options:           o,
		TagSpecifications: tspecs,
	}

	return nm.ClientEC2[*region].CreateTransitGateway(nm.ctx, input)
}

// DescribeTransitGateways function
func (nm *NMgr) DescribeTransitGateways(region, name *string) (*ec2.DescribeTransitGatewaysOutput, error) {
	tagKey := "tag:Name"
	filters := createEC2Filter(&tagKey, name)

	input := &ec2.DescribeTransitGatewaysInput{
		Filters: filters,
	}
	return nm.ClientEC2[*region].DescribeTransitGateways(nm.ctx, input)
}

// DeleteTransitGateway function
func (nm *NMgr) DeleteTransitGateway(region, id *string) (*ec2.DeleteTransitGatewayOutput, error) {
	input := &ec2.DeleteTransitGatewayInput{
		TransitGatewayId: id,
	}
	return nm.ClientEC2[*region].DeleteTransitGateway(nm.ctx, input)
}

// CreateCustomerGateway fucntion
func (nm *NMgr) CreateCustomerGateway(region, name, ip *string, asn *int32) (*ec2.CreateCustomerGatewayOutput, error) {
	r, err := nm.DescribeCustomerGateways(region, name)
	if err != nil {
		log.Fatal(err)
	}
	if len(r.CustomerGateways) > 0 {
		// TransitGateway exists
		log.Infof("Customer Gateway exists")
		o := &ec2.CreateCustomerGatewayOutput{
			CustomerGateway: &r.CustomerGateways[0],
		}
		return o, nil
	}

	tagKey := "Name"
	tspecs := createEC2TagSpecs(&tagKey, name, types.ResourceTypeCustomerGateway)

	input := &ec2.CreateCustomerGatewayInput{
		BgpAsn:            *asn,
		Type:              types.GatewayTypeIpsec1,
		DeviceName:        name,
		PublicIp:          ip,
		TagSpecifications: tspecs,
	}
	return nm.ClientEC2[*region].CreateCustomerGateway(nm.ctx, input)
}

// DescribeCustomerGateways function
func (nm *NMgr) DescribeCustomerGateways(region, name *string) (*ec2.DescribeCustomerGatewaysOutput, error) {
	tagKey := "tag:Name"
	filters := createEC2Filter(&tagKey, name)

	input := &ec2.DescribeCustomerGatewaysInput{
		Filters: filters,
	}
	return nm.ClientEC2[*region].DescribeCustomerGateways(nm.ctx, input)
}

// DeleteCustomerGateway function
func (nm *NMgr) DeleteCustomerGateway(region, id *string) (*ec2.DeleteCustomerGatewayOutput, error) {
	input := &ec2.DeleteCustomerGatewayInput{
		CustomerGatewayId: id,
	}
	return nm.ClientEC2[*region].DeleteCustomerGateway(nm.ctx, input)
}

// CreateVpnConnection function
func (nm *NMgr) CreateVpnConnection(region, name, cgwID, tgwID, cidr *string) (*ec2.CreateVpnConnectionOutput, error) {
	r, err := nm.DescribeVpnConnections(region, name)
	if err != nil {
		log.Fatal(err)
	}
	if len(r.VpnConnections) > 0 {
		// VPN connection exists
		log.Infof("VPN connection exists")
		o := &ec2.CreateVpnConnectionOutput{
			VpnConnection: &r.VpnConnections[0],
		}
		return o, nil
	}

	tagKey := "Name"
	tspecs := createEC2TagSpecs(&tagKey, name, types.ResourceTypeVpnConnection)

	tunnelOption := types.VpnTunnelOptionsSpecification{
		PreSharedKey: &psk,
	}

	var tunnelOptions []types.VpnTunnelOptionsSpecification
	tunnelOptions = append(tunnelOptions, tunnelOption)

	options := &types.VpnConnectionOptionsSpecification{
		LocalIpv4NetworkCidr: cidr,
		StaticRoutesOnly:     true,
		TunnelOptions:        tunnelOptions,
	}

	tgwType := "ipsec.1"

	input := &ec2.CreateVpnConnectionInput{
		CustomerGatewayId: cgwID,
		Type:              &tgwType,
		TransitGatewayId:  tgwID,
		TagSpecifications: tspecs,
		Options:           options,
	}
	return nm.ClientEC2[*region].CreateVpnConnection(nm.ctx, input)
}

// DescribeVpnConnections function
func (nm *NMgr) DescribeVpnConnections(region, name *string) (*ec2.DescribeVpnConnectionsOutput, error) {
	tagKey := "tag:Name"
	filters := createEC2Filter(&tagKey, name)

	input := &ec2.DescribeVpnConnectionsInput{
		Filters: filters,
	}
	return nm.ClientEC2[*region].DescribeVpnConnections(nm.ctx, input)
}

// DeleteVpnConnection function
func (nm *NMgr) DeleteVpnConnection(region, id *string) (*ec2.DeleteVpnConnectionOutput, error) {
	input := &ec2.DeleteVpnConnectionInput{
		VpnConnectionId: id,
	}
	return nm.ClientEC2[*region].DeleteVpnConnection(nm.ctx, input)
}
