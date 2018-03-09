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
}

func NewDBInstance(client dbInstancesClient, name *string) DBInstance {
	return DBInstance{
		client:     client,
		name:       name,
		identifier: *name,
	}
}

func (d DBInstance) Delete() error {
	_, err := d.client.DeleteDBInstance(&awsrds.DeleteDBInstanceInput{
		DBInstanceIdentifier: d.name,
		SkipFinalSnapshot:    aws.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting db instance %s: %s", d.identifier, err)
	}

	return nil
}

func (d DBInstance) Name() string {
	return d.identifier
}

func (d DBInstance) Type() string {
	return "db instance"
}
