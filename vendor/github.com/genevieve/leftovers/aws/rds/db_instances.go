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

func (d DBInstances) List(filter string) ([]common.Deletable, error) {
	dbInstances, err := d.client.DescribeDBInstances(&awsrds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing db instances: %s", err)
	}

	var resources []common.Deletable
	for _, db := range dbInstances.DBInstances {
		resource := NewDBInstance(d.client, db.DBInstanceIdentifier)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := d.logger.Prompt(fmt.Sprintf("Are you sure you want to delete db instance %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
