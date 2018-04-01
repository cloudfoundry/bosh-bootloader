package rds

import (
	"fmt"
	"strings"

	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/genevieve/leftovers/aws/common"
)

type dbSubnetGroupsClient interface {
	DescribeDBSubnetGroups(*awsrds.DescribeDBSubnetGroupsInput) (*awsrds.DescribeDBSubnetGroupsOutput, error)
	DeleteDBSubnetGroup(*awsrds.DeleteDBSubnetGroupInput) (*awsrds.DeleteDBSubnetGroupOutput, error)
}

type DBSubnetGroups struct {
	client dbSubnetGroupsClient
	logger logger
}

func NewDBSubnetGroups(client dbSubnetGroupsClient, logger logger) DBSubnetGroups {
	return DBSubnetGroups{
		client: client,
		logger: logger,
	}
}

func (d DBSubnetGroups) ListOnly(filter string) ([]common.Deletable, error) {
	return d.get(filter)
}

func (d DBSubnetGroups) List(filter string) ([]common.Deletable, error) {
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

func (d DBSubnetGroups) get(filter string) ([]common.Deletable, error) {
	dbSubnetGroups, err := d.client.DescribeDBSubnetGroups(&awsrds.DescribeDBSubnetGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing RDS DB Subnet Groups: %s", err)
	}

	var resources []common.Deletable
	for _, db := range dbSubnetGroups.DBSubnetGroups {
		resource := NewDBSubnetGroup(d.client, db.DBSubnetGroupName)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
