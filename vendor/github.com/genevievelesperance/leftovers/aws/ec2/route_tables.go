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
	client routesClient
	logger logger
}

func NewRouteTables(client routesClient, logger logger) RouteTables {
	return RouteTables{
		client: client,
		logger: logger,
	}
}

func (u RouteTables) Delete(vpcId string) error {
	routeTables, err := u.client.DescribeRouteTables(&awsec2.DescribeRouteTablesInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("vpc-id"),
			Values: []*string{aws.String(vpcId)},
		}},
	})
	if err != nil {
		return fmt.Errorf("Describing route tables: %s", err)
	}

	for _, r := range routeTables.RouteTables {
		n := *r.RouteTableId

		if len(r.Associations) > 0 {
			isMain := *r.Associations[0].Main
			if isMain {
				continue
			}
		}

		for _, a := range r.Associations {
			_, err = u.client.DisassociateRouteTable(&awsec2.DisassociateRouteTableInput{
				AssociationId: a.RouteTableAssociationId,
			})
			if err == nil {
				u.logger.Printf("SUCCESS disassociating route table %s\n", n)
			} else {
				u.logger.Printf("ERROR disassociating route table %s: %s\n", n, err)
			}
		}

		_, err = u.client.DeleteRouteTable(&awsec2.DeleteRouteTableInput{
			RouteTableId: r.RouteTableId,
		})
		if err == nil {
			u.logger.Printf("SUCCESS deleting route table %s\n", n)
		} else {
			u.logger.Printf("ERROR deleting route table %s: %s\n", n, err)
		}
	}

	return nil

}
