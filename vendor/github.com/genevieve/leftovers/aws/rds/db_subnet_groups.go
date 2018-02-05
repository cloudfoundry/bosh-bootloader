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

func (d DBSubnetGroups) List(filter string) ([]common.Deletable, error) {
	dbSubnetGroups, err := d.client.DescribeDBSubnetGroups(&awsrds.DescribeDBSubnetGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing db subnet groups: %s", err)
	}

	var resources []common.Deletable
	for _, db := range dbSubnetGroups.DBSubnetGroups {
		resource := NewDBSubnetGroup(d.client, db.DBSubnetGroupName)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := d.logger.Prompt(fmt.Sprintf("Are you sure you want to delete db subnet group %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
