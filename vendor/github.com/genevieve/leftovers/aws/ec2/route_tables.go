package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type routesClient interface {
	DescribeRouteTables(*awsec2.DescribeRouteTablesInput) (*awsec2.DescribeRouteTablesOutput, error)
	DisassociateRouteTable(*awsec2.DisassociateRouteTableInput) (*awsec2.DisassociateRouteTableOutput, error)
	DeleteRouteTable(*awsec2.DeleteRouteTableInput) (*awsec2.DeleteRouteTableOutput, error)
}

type routeTables interface {
	Delete(vpcId string) error
}

type RouteTables struct {
	client       routesClient
	logger       logger
	resourceTags resourceTags
}

func NewRouteTables(client routesClient, logger logger, resourceTags resourceTags) RouteTables {
	return RouteTables{
		client:       client,
		logger:       logger,
		resourceTags: resourceTags,
	}
}

func (u RouteTables) Delete(vpcId string) error {
	routeTables, err := u.client.DescribeRouteTables(&awsec2.DescribeRouteTablesInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("vpc-id"),
			Values: []*string{aws.String(vpcId)},
		}, {
			Name:   aws.String("association.main"),
			Values: []*string{aws.String("false")},
		}},
	})
	if err != nil {
		return fmt.Errorf("Describe EC2 Route Tables: %s", err)
	}

	for _, r := range routeTables.RouteTables {
		n := *r.RouteTableId

		for _, a := range r.Associations {
			_, err = u.client.DisassociateRouteTable(&awsec2.DisassociateRouteTableInput{AssociationId: a.RouteTableAssociationId})
			if err == nil {
				u.logger.Printf("[EC2 VPC: %s] Disassociated route table %s", vpcId, n)
			} else {
				u.logger.Printf("[EC2 VPC: %s] Disassociate route table %s: %s", vpcId, n, err)
			}
		}

		_, err = u.client.DeleteRouteTable(&awsec2.DeleteRouteTableInput{RouteTableId: r.RouteTableId})
		if err != nil {
			return fmt.Errorf("Delete %s: %s", n, err)
		} else {
			u.logger.Printf("[EC2 VPC: %s] Deleted route table %s", vpcId, n)
		}

		err = u.resourceTags.Delete("route-table", n)
		if err != nil {
			u.logger.Printf("[EC2 VPC: %s] Delete route table %s tags: %s", vpcId, n, err)
		} else {
			u.logger.Printf("[EC2 VPC: %s] Deleted route table %s tags", vpcId, n)
		}
	}

	return nil
}
