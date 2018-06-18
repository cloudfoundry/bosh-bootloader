package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type internetGatewaysClient interface {
	DescribeInternetGateways(*awsec2.DescribeInternetGatewaysInput) (*awsec2.DescribeInternetGatewaysOutput, error)
	DetachInternetGateway(*awsec2.DetachInternetGatewayInput) (*awsec2.DetachInternetGatewayOutput, error)
	DeleteInternetGateway(*awsec2.DeleteInternetGatewayInput) (*awsec2.DeleteInternetGatewayOutput, error)
}

type internetGateways interface {
	Delete(vpcId string) error
}

type InternetGateways struct {
	client internetGatewaysClient
	logger logger
}

func NewInternetGateways(client internetGatewaysClient, logger logger) InternetGateways {
	return InternetGateways{
		client: client,
		logger: logger,
	}
}

func (n InternetGateways) Delete(vpcId string) error {
	igws, err := n.client.DescribeInternetGateways(&awsec2.DescribeInternetGatewaysInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("attachment.vpc-id"),
			Values: []*string{aws.String(vpcId)},
		}},
	})
	if err != nil {
		return fmt.Errorf("Describe EC2 Internet Gateways: %s", err)
	}

	for _, i := range igws.InternetGateways {
		igwId := *i.InternetGatewayId

		_, err = n.client.DetachInternetGateway(&awsec2.DetachInternetGatewayInput{
			InternetGatewayId: i.InternetGatewayId,
			VpcId:             aws.String(vpcId),
		})
		if err != nil {
			n.logger.Printf("[EC2 VPC: %s] Detach internet gateway %s: %s \n", vpcId, igwId, err)
		} else {
			n.logger.Printf("[EC2 VPC: %s] Detached internet gateway %s \n", vpcId, igwId)
		}

		_, err = n.client.DeleteInternetGateway(&awsec2.DeleteInternetGatewayInput{
			InternetGatewayId: i.InternetGatewayId,
		})
		if err != nil {
			return fmt.Errorf("Delete %s: %s", igwId, err)
		} else {
			n.logger.Printf("[EC2 VPC: %s] Deleted internet gateway %s \n", vpcId, igwId)
		}
	}

	return nil
}
