package awsnmgr

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/kelvins/geocoder"
	log "github.com/sirupsen/logrus"
)

func createNetwTags(tagKey, tagValue *string) (tags []types.Tag) {
	tag := types.Tag{
		Key:   tagKey,
		Value: tagValue,
	}
	tags = append(tags, tag)
	return tags
}

// CreateGlobalNetwork function
func (nm *NMgr) CreateGlobalNetwork(name *string) (*networkmanager.CreateGlobalNetworkOutput, error) {
	r, err := nm.DescribeGlobalNetworks()
	if err != nil {
		log.Fatal(err)
	}

	//if len(r.GlobalNetworks) > 0 {
	for idx, g := range r.GlobalNetworks {
		for i := 0; i < len(g.Tags); i++ {
			if *g.Tags[i].Key == "Name" {
				if *g.Tags[i].Value == *name {
					log.Infof("Global Betwork exists")
					o := &networkmanager.CreateGlobalNetworkOutput{
						GlobalNetwork: &r.GlobalNetworks[idx],
					}
					return o, nil
				}
			}
		}
	}
	//}

	tagKey := "Name"
	tags := createNetwTags(&tagKey, name)

	input := &networkmanager.CreateGlobalNetworkInput{
		Description: name,
		Tags:        tags,
	}

	return nm.ClientNMgr.CreateGlobalNetwork(nm.ctx, input)
}

// DeleteGlobalNetwork function
func (nm *NMgr) DeleteGlobalNetwork() (*networkmanager.DeleteGlobalNetworkOutput, error) {
	input := &networkmanager.DeleteGlobalNetworkInput{
		GlobalNetworkId: nm.GlobalNetworkID,
	}
	return nm.ClientNMgr.DeleteGlobalNetwork(nm.ctx, input)
}

// CreateSite function
func (nm *NMgr) CreateSite(name *string, s *Site) (*networkmanager.CreateSiteOutput, error) {
	r, err := nm.GetSites()
	if err != nil {
		log.Fatal(err)
	}
	for idx, g := range r.Sites {
		for i := 0; i < len(g.Tags); i++ {
			if *g.Tags[i].Key == "Name" {
				if *g.Tags[i].Value == *name {
					log.Infof("Site exists")
					o := &networkmanager.CreateSiteOutput{
						Site: &r.Sites[idx],
					}
					return o, nil
				}
			}
		}
	}

	tagKey := "Name"
	tags := createNetwTags(&tagKey, name)

	address := geocoder.Address{
		Street:  s.Street,
		Number:  s.Number,
		City:    s.City,
		State:   s.State,
		Country: s.Country,
	}

	a := address.Street + ", " + fmt.Sprintf("%d", address.Number) + ", " + address.City + ", " + address.State + ", " + address.Country
	location := &types.Location{
		Address: &a,
	}

	input := &networkmanager.CreateSiteInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		Description:     name,
		Location:        location,
		Tags:            tags,
	}

	return nm.ClientNMgr.CreateSite(nm.ctx, input)

}

// DeleteSite function
func (nm *NMgr) DeleteSite(s *string) (*networkmanager.DeleteSiteOutput, error) {
	input := &networkmanager.DeleteSiteInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		SiteId:          s,
	}
	return nm.ClientNMgr.DeleteSite(nm.ctx, input)
}

// DeleteSites function
func (nm *NMgr) DeleteSites() error {
	r, err := nm.GetSites()
	if err != nil {
		log.Error(err)
	}
	for _, s := range r.Sites {
		log.Infof("Delete site: %s", *s.SiteId)
		_, err := nm.DeleteSite(s.SiteId)
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

// CreateDevice function
func (nm *NMgr) CreateDevice(name *string, d *Device) (*networkmanager.CreateDeviceOutput, error) {
	r, err := nm.GetDevices()
	if err != nil {
		log.Fatal(err)
	}
	for idx, g := range r.Devices {
		for i := 0; i < len(g.Tags); i++ {
			if *g.Tags[i].Key == "Name" {
				if *g.Tags[i].Value == *name {
					log.Infof("Site exists")
					o := &networkmanager.CreateDeviceOutput{
						Device: &r.Devices[idx],
					}
					return o, nil
				}
			}
		}
	}

	tagKey := "Name"
	tags := createNetwTags(&tagKey, name)

	address := geocoder.Address{
		Street:  d.Site.Street,
		Number:  d.Site.Number,
		City:    d.Site.City,
		State:   d.Site.State,
		Country: d.Site.Country,
	}

	a := address.Street + ", " + fmt.Sprintf("%d", address.Number) + ", " + address.City + ", " + address.State + ", " + address.Country
	location := &types.Location{
		Address:   &a,
	}
	model := d.Model
	serial := d.Serial
	dtype := d.Kind
	vendor := "nuage"

	input := &networkmanager.CreateDeviceInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		Description:     name,
		Location:        location,
		Model:           &model,
		SerialNumber:    &serial,
		Type:            &dtype,
		Vendor:          &vendor,
		Tags:            tags,
		SiteId:          d.Site.SiteID,
	}

	return nm.ClientNMgr.CreateDevice(nm.ctx, input)
}

// DeleteDevice function
func (nm *NMgr) DeleteDevice(d *string) (*networkmanager.DeleteDeviceOutput, error) {
	input := &networkmanager.DeleteDeviceInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		DeviceId:        d,
	}
	return nm.ClientNMgr.DeleteDevice(nm.ctx, input)
}

// DeleteDevices function
func (nm *NMgr) DeleteDevices() error {
	r, err := nm.GetDevices()
	if err != nil {
		log.Error(err)
	}
	for _, d := range r.Devices {
		log.Infof("Delete device: %s", *d.DeviceId)
		_, err := nm.DeleteDevice(d.DeviceId)
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

// CreateLink function
func (nm *NMgr) CreateLink(name *string, ep *Endpoint) (*networkmanager.CreateLinkOutput, error) {
	r, err := nm.GetLinks()
	if err != nil {
		log.Fatal(err)
	}
	//if len(r.Links) > 0 {
	for idx, g := range r.Links {
		for i := 0; i < len(g.Tags); i++ {
			if *g.Tags[i].Key == "Name" {
				if *g.Tags[i].Value == *name {
					log.Infof("Site exists")
					o := &networkmanager.CreateLinkOutput{
						Link: &r.Links[idx],
					}
					return o, nil
				}
			}
		}
	}
	//}

	tagKey := "Name"
	tags := createNetwTags(&tagKey, name)

	bw := &types.Bandwidth{
		DownloadSpeed: &ep.BwDown,
		UploadSpeed:   &ep.BwUp,
	}

	input := &networkmanager.CreateLinkInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		SiteId:          ep.Site.SiteID,
		Bandwidth:       bw,
		Description:     name,
		Provider:        &ep.Provider,
		Type:            &ep.Kind,
		Tags:            tags,
	}

	return nm.ClientNMgr.CreateLink(nm.ctx, input)
}

// DeleteLink function
func (nm *NMgr) DeleteLink(l *string) (*networkmanager.DeleteLinkOutput, error) {
	input := &networkmanager.DeleteLinkInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		LinkId:          l,
	}
	return nm.ClientNMgr.DeleteLink(nm.ctx, input)
}

// DeleteLinks function
func (nm *NMgr) DeleteLinks() error {
	r, err := nm.GetLinks()
	if err != nil {
		log.Error(err)
	}
	for _, l := range r.Links {
		log.Infof("Delete link: %s", *l.LinkId)
		_, err := nm.DeleteLink(l.LinkId)
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

// DescribeGlobalNetworks function
func (nm *NMgr) DescribeGlobalNetworks() (*networkmanager.DescribeGlobalNetworksOutput, error) {
	input := &networkmanager.DescribeGlobalNetworksInput{}
	return nm.ClientNMgr.DescribeGlobalNetworks(nm.ctx, input)
}

// GetSites function
func (nm *NMgr) GetSites() (*networkmanager.GetSitesOutput, error) {
	input := &networkmanager.GetSitesInput{
		GlobalNetworkId: nm.GlobalNetworkID,
	}
	return nm.ClientNMgr.GetSites(nm.ctx, input)
}

// GetDevices function
func (nm *NMgr) GetDevices() (*networkmanager.GetDevicesOutput, error) {
	input := &networkmanager.GetDevicesInput{
		GlobalNetworkId: nm.GlobalNetworkID,
	}
	return nm.ClientNMgr.GetDevices(nm.ctx, input)
}

// GetDevice function
func (nm *NMgr) GetDevice(siteID *string) (*networkmanager.GetDevicesOutput, error) {
	input := &networkmanager.GetDevicesInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		SiteId:          siteID,
	}
	return nm.ClientNMgr.GetDevices(nm.ctx, input)
}

// GetLinks function
func (nm *NMgr) GetLinks() (*networkmanager.GetLinksOutput, error) {
	input := &networkmanager.GetLinksInput{
		GlobalNetworkId: nm.GlobalNetworkID,
	}
	return nm.ClientNMgr.GetLinks(nm.ctx, input)
}

// GetLink function
func (nm *NMgr) GetLink(siteID *string) (*networkmanager.GetLinksOutput, error) {
	input := &networkmanager.GetLinksInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		SiteId:          siteID,
	}
	return nm.ClientNMgr.GetLinks(nm.ctx, input)
}

// RegisterTransitGateway function
func (nm *NMgr) RegisterTransitGateway(arn *string) (*networkmanager.RegisterTransitGatewayOutput, error) {
	input := &networkmanager.RegisterTransitGatewayInput{
		GlobalNetworkId:   nm.GlobalNetworkID,
		TransitGatewayArn: arn,
	}
	return nm.ClientNMgr.RegisterTransitGateway(nm.ctx, input)
}

// DeregisterTransitGateway function
func (nm *NMgr) DeregisterTransitGateway(arn *string) (*networkmanager.DeregisterTransitGatewayOutput, error) {
	input := &networkmanager.DeregisterTransitGatewayInput{
		GlobalNetworkId:   nm.GlobalNetworkID,
		TransitGatewayArn: arn,
	}
	return nm.ClientNMgr.DeregisterTransitGateway(nm.ctx, input)
}

// GetTransitGatewayRegistrations fucntion
func (nm *NMgr) GetTransitGatewayRegistrations() (*networkmanager.GetTransitGatewayRegistrationsOutput, error) {
	input := &networkmanager.GetTransitGatewayRegistrationsInput{
		GlobalNetworkId:   nm.GlobalNetworkID,
	}
	return nm.ClientNMgr.GetTransitGatewayRegistrations(nm.ctx, input)
}

// AssociateCustomerGateway function
func (nm *NMgr) AssociateCustomerGateway(cgwArn, dID, lID *string) (*networkmanager.AssociateCustomerGatewayOutput, error) {
	input := &networkmanager.AssociateCustomerGatewayInput{
		GlobalNetworkId:    nm.GlobalNetworkID,
		CustomerGatewayArn: cgwArn,
		DeviceId:           dID,
		LinkId:             lID,
	}
	return nm.ClientNMgr.AssociateCustomerGateway(nm.ctx, input)
}

// DisassociateCustomerGateway function
func (nm *NMgr) DisassociateCustomerGateway(cgwArn, dID, lID *string) (*networkmanager.DisassociateCustomerGatewayOutput, error) {
	input := &networkmanager.DisassociateCustomerGatewayInput{
		GlobalNetworkId:    nm.GlobalNetworkID,
		CustomerGatewayArn: cgwArn,
	}
	return nm.ClientNMgr.DisassociateCustomerGateway(nm.ctx, input)
}

// AssociateLink function
func (nm *NMgr) AssociateLink(dID, lID *string) (*networkmanager.AssociateLinkOutput, error) {
	input := &networkmanager.AssociateLinkInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		DeviceId:        dID,
		LinkId:          lID,
	}
	return nm.ClientNMgr.AssociateLink(nm.ctx, input)
}

// DisassociateLink function
func (nm *NMgr) DisassociateLink(dID, lID *string) (*networkmanager.DisassociateLinkOutput, error) {
	input := &networkmanager.DisassociateLinkInput{
		GlobalNetworkId: nm.GlobalNetworkID,
		DeviceId:        dID,
		LinkId:          lID,
	}
	return nm.ClientNMgr.DisassociateLink(nm.ctx, input)
}

// GetCustomerGatewayAssociations function
func (nm *NMgr) GetCustomerGatewayAssociations() (*networkmanager.GetCustomerGatewayAssociationsOutput, error) {
	input := &networkmanager.GetCustomerGatewayAssociationsInput{
		GlobalNetworkId: nm.GlobalNetworkID,
	}
	return nm.ClientNMgr.GetCustomerGatewayAssociations(nm.ctx, input)
}
