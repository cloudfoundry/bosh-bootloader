package rds

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
)

type DBInstance struct {
	client     dbInstancesClient
	name       *string
	identifier string
	rtype      string
}

func NewDBInstance(client dbInstancesClient, name *string) DBInstance {
	return DBInstance{
		client:     client,
		name:       name,
		identifier: *name,
		rtype:      "RDS DB Instance",
	}
}

func (d DBInstance) Delete() error {
	_, err := d.client.DeleteDBInstance(&awsrds.DeleteDBInstanceInput{
		DBInstanceIdentifier: d.name,
		SkipFinalSnapshot:    aws.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (d DBInstance) Name() string {
	return d.identifier
}

func (d DBInstance) Type() string {
	return d.rtype
}
