package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type subnetsClient interface {
	DescribeSubnets(*awsec2.DescribeSubnetsInput) (*awsec2.DescribeSubnetsOutput, error)
	DeleteSubnet(*awsec2.DeleteSubnetInput) (*awsec2.DeleteSubnetOutput, error)
}

type subnets interface {
	Delete(vpcId string) error
}

type Subnets struct {
	client subnetsClient
	logger logger
}

func NewSubnets(client subnetsClient, logger logger) Subnets {
	return Subnets{
		client: client,
		logger: logger,
	}
}

func (u Subnets) Delete(vpcId string) error {
	subnets, err := u.client.DescribeSubnets(&awsec2.DescribeSubnetsInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("vpc-id"),
			Values: []*string{aws.String(vpcId)},
		}},
	})
	if err != nil {
		return fmt.Errorf("Describing subnets: %s", err)
	}

	for _, s := range subnets.Subnets {
		n := *s.SubnetId

		_, err = u.client.DeleteSubnet(&awsec2.DeleteSubnetInput{SubnetId: s.SubnetId})

		if err == nil {
			u.logger.Printf("SUCCESS deleting subnet %s\n", n)
		} else {
			u.logger.Printf("ERROR deleting subnet %s: %s\n", n, err)
		}
	}

	return nil
}
