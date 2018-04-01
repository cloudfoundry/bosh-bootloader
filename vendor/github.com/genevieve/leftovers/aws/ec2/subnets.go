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
	client       subnetsClient
	logger       logger
	resourceTags resourceTags
}

func NewSubnets(client subnetsClient, logger logger, resourceTags resourceTags) Subnets {
	return Subnets{
		client:       client,
		logger:       logger,
		resourceTags: resourceTags,
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
		return fmt.Errorf("Describe EC2 Subnets: %s", err)
	}

	for _, s := range subnets.Subnets {
		n := *s.SubnetId

		_, err = u.client.DeleteSubnet(&awsec2.DeleteSubnetInput{SubnetId: s.SubnetId})
		if err != nil {
			return fmt.Errorf("Delete subnet %s: %s", n, err)
		}

		err = u.resourceTags.Delete("subnet", n)
		if err != nil {
			u.logger.Printf("[EC2 VPC: %s] Delete subnet %s tags: %s", vpcId, n, err)
		} else {
			u.logger.Printf("[EC2 VPC: %s] Deleted subnet %s tags", vpcId, n)
		}
	}

	return nil
}
