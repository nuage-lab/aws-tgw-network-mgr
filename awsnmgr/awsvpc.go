package awsnmgr

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
)

// CreateVpc function
func (nm *NMgr) CreateVpc(region, name, cidr *string) (*ec2.CreateVpcOutput, error) {

	r, err := nm.DescribeVpcs(region, name)
	if err != nil {
		log.Fatal(err)
	}

	if len(r.Vpcs) > 0 {
		log.Infof("VPC exists")
		o := &ec2.CreateVpcOutput{
			Vpc: &r.Vpcs[0],
		}
		return o, nil
	}

	tagKey := "Name"
	tspecs := createEC2TagSpecs(&tagKey, name, types.ResourceTypeVpc)

	input := &ec2.CreateVpcInput{
		CidrBlock:         cidr,
		TagSpecifications: tspecs,
	}

	return nm.ClientEC2[*region].CreateVpc(nm.ctx, input)
}

// DescribeVpcs function
func (nm *NMgr) DescribeVpcs(region, name *string) (*ec2.DescribeVpcsOutput, error) {
	tagKey := "tag:Name"
	filters := createEC2Filter(&tagKey, name)

	input := &ec2.DescribeVpcsInput{
		Filters: filters,
	}
	return nm.ClientEC2[*region].DescribeVpcs(nm.ctx, input)
}
