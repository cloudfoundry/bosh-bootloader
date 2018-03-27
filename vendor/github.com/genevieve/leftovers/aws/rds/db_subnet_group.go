package rds

import (
	"fmt"

	awsrds "github.com/aws/aws-sdk-go/service/rds"
)

type DBSubnetGroup struct {
	client     dbSubnetGroupsClient
	name       *string
	identifier string
	rtype      string
}

func NewDBSubnetGroup(client dbSubnetGroupsClient, name *string) DBSubnetGroup {
	return DBSubnetGroup{
		client:     client,
		name:       name,
		identifier: *name,
		rtype:      "RDS DB Subnet Group",
	}
}

func (d DBSubnetGroup) Delete() error {
	_, err := d.client.DeleteDBSubnetGroup(&awsrds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: d.name,
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting %s %s: %s", d.rtype, d.identifier, err)
	}

	return nil
}

func (d DBSubnetGroup) Name() string {
	return d.identifier
}

func (d DBSubnetGroup) Type() string {
	return d.rtype
}
