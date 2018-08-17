package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/common"
)

type vpcsClient interface {
	DescribeVpcs(*awsec2.DescribeVpcsInput) (*awsec2.DescribeVpcsOutput, error)
	DeleteVpc(*awsec2.DeleteVpcInput) (*awsec2.DeleteVpcOutput, error)
}

type Vpcs struct {
	client       vpcsClient
	logger       logger
	routes       routeTables
	subnets      subnets
	gateways     internetGateways
	resourceTags resourceTags
}

func NewVpcs(client vpcsClient, logger logger, routes routeTables, subnets subnets, gateways internetGateways, resourceTags resourceTags) Vpcs {
	return Vpcs{
		client:       client,
		logger:       logger,
		routes:       routes,
		subnets:      subnets,
		gateways:     gateways,
		resourceTags: resourceTags,
	}
}

func (v Vpcs) List(filter string) ([]common.Deletable, error) {
	output, err := v.client.DescribeVpcs(&awsec2.DescribeVpcsInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("isDefault"),
			Values: []*string{aws.String("false")},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("Describe EC2 VPCs: %s", err)
	}

	var resources []common.Deletable
	for _, vpc := range output.Vpcs {
		r := NewVpc(v.client, v.routes, v.subnets, v.gateways, v.resourceTags, vpc.VpcId, vpc.Tags)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := v.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (v Vpcs) Type() string {
	return "ec2-vpc"
}
