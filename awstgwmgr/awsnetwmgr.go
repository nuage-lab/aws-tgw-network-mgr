package awsnmgr

import (
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
)

// CreateGlobalNetwork function
func (c *NMgr) CreateGlobalNetwork(n *string) (*networkmanager.CreateGlobalNetworkOutput, error) {
	tagKey := "name"
	
	tag := types.Tag{
		Key: &tagKey,
		Value: n,
	}
	var tags []types.Tag
	tags = append(tags, tag)

	input := &networkmanager.CreateGlobalNetworkInput{
		Description: n,
		Tags:        tags,
	}

	return c.ClientNMgr.CreateGlobalNetwork(c.ctx, input)
}
