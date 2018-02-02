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
		return fmt.Errorf("Describing internet gateways: %s", err)
	}

	for _, i := range igws.InternetGateways {
		igwId := *i.InternetGatewayId

		_, err = n.client.DetachInternetGateway(&awsec2.DetachInternetGatewayInput{
			InternetGatewayId: i.InternetGatewayId,
			VpcId:             aws.String(vpcId),
		})
		if err == nil {
			n.logger.Printf("SUCCESS detaching internet gateway %s\n", igwId)
		} else {
			n.logger.Printf("ERROR detaching internet gateway %s: %s\n", igwId, err)
		}

		_, err = n.client.DeleteInternetGateway(&awsec2.DeleteInternetGatewayInput{
			InternetGatewayId: i.InternetGatewayId,
		})
		if err == nil {
			n.logger.Printf("SUCCESS deleting internet gateway %s\n", igwId)
		} else {
			n.logger.Printf("ERROR deleting internet gateway %s: %s\n", igwId, err)
		}
	}

	return nil
}
