package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
)

type vpcsClient interface {
	DescribeVpcs(*awsec2.DescribeVpcsInput) (*awsec2.DescribeVpcsOutput, error)
	DeleteVpc(*awsec2.DeleteVpcInput) (*awsec2.DeleteVpcOutput, error)
}

type Vpcs struct {
	client   vpcsClient
	logger   logger
	routes   routeTables
	subnets  subnets
	gateways internetGateways
}

func NewVpcs(client vpcsClient,
	logger logger,
	routes routeTables,
	subnets subnets,
	gateways internetGateways) Vpcs {
	return Vpcs{
		client:   client,
		logger:   logger,
		routes:   routes,
		subnets:  subnets,
		gateways: gateways,
	}
}

func (v Vpcs) ListAll(filter string) ([]common.Deletable, error) {
	return v.get(filter)
}

func (v Vpcs) List(filter string) ([]common.Deletable, error) {
	resources, err := v.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := v.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (v Vpcs) get(filter string) ([]common.Deletable, error) {
	output, err := v.client.DescribeVpcs(&awsec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing EC2 VPCs: %s", err)
	}

	var resources []common.Deletable
	for _, vpc := range output.Vpcs {
		resource := NewVpc(v.client, v.routes, v.subnets, v.gateways, vpc.VpcId, vpc.Tags)

		if *vpc.IsDefault {
			continue
		}

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
