package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/common"
)

type natGatewaysClient interface {
	DescribeNatGateways(*awsec2.DescribeNatGatewaysInput) (*awsec2.DescribeNatGatewaysOutput, error)
	DeleteNatGateway(*awsec2.DeleteNatGatewayInput) (*awsec2.DeleteNatGatewayOutput, error)
}

type NatGateways struct {
	client natGatewaysClient
	logger logger
}

func NewNatGateways(client natGatewaysClient, logger logger) NatGateways {
	return NatGateways{
		client: client,
		logger: logger,
	}
}

func (n NatGateways) List(filter string) ([]common.Deletable, error) {
	natGateways, err := n.client.DescribeNatGateways(&awsec2.DescribeNatGatewaysInput{
		Filter: []*awsec2.Filter{{
			Name:   aws.String("state"),
			Values: []*string{aws.String("pending"), aws.String("available")},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("Describing EC2 Nat Gateways: %s", err)
	}

	var resources []common.Deletable
	for _, g := range natGateways.NatGateways {
		r := NewNatGateway(n.client, n.logger, g.NatGatewayId, g.Tags)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := n.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (n NatGateways) Type() string {
	return "ec2-nat-gateway"
}
