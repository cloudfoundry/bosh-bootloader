package rds

import (
	"fmt"
	"strings"

	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/genevieve/leftovers/aws/common"
)

type dbInstancesClient interface {
	DescribeDBInstances(*awsrds.DescribeDBInstancesInput) (*awsrds.DescribeDBInstancesOutput, error)
	DeleteDBInstance(*awsrds.DeleteDBInstanceInput) (*awsrds.DeleteDBInstanceOutput, error)
}

type DBInstances struct {
	client dbInstancesClient
	logger logger
}

func NewDBInstances(client dbInstancesClient, logger logger) DBInstances {
	return DBInstances{
		client: client,
		logger: logger,
	}
}

func (d DBInstances) ListOnly(filter string) ([]common.Deletable, error) {
	return d.get(filter)
}

func (d DBInstances) List(filter string) ([]common.Deletable, error) {
	resources, err := d.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := d.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (d DBInstances) get(filter string) ([]common.Deletable, error) {
	dbInstances, err := d.client.DescribeDBInstances(&awsrds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing RDS DB Instances: %s", err)
	}

	var resources []common.Deletable
	for _, db := range dbInstances.DBInstances {
		resource := NewDBInstance(d.client, db.DBInstanceIdentifier)

		if *db.DBInstanceStatus == "deleting" {
			continue
		}

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
