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
	rtype  string
}

func NewSubnets(client subnetsClient, logger logger) Subnets {
	return Subnets{
		client: client,
		logger: logger,
		rtype:  "EC2 Subnet",
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
		return fmt.Errorf("Describing EC2 Subnets: %s", err)
	}

	for _, s := range subnets.Subnets {
		n := *s.SubnetId

		_, err = u.client.DeleteSubnet(&awsec2.DeleteSubnetInput{SubnetId: s.SubnetId})

		if err == nil {
			u.logger.Printf("SUCCESS deleting %s %s\n", u.rtype, n)
		} else {
			u.logger.Printf("ERROR deleting %s %s: %s\n", u.rtype, n, err)
		}
	}

	return nil
}
