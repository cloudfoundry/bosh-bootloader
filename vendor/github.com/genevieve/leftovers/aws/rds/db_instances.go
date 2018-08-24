package rds

import (
	"fmt"
	"strings"

	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/genevieve/leftovers/common"
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
		return nil, fmt.Errorf("Describing RDS DB Instances: %s", err)
	}

	var resources []common.Deletable
	for _, db := range dbInstances.DBInstances {
		if *db.DBInstanceStatus == "deleting" {
			continue
		}

		r := NewDBInstance(d.client, db.DBInstanceIdentifier)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := d.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (d DBInstances) Type() string {
	return "rds-db-instance"
}
