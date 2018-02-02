package elbv2

import (
	"fmt"
	"strings"

	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/genevievelesperance/leftovers/aws/common"
)

type targetGroupsClient interface {
	DescribeTargetGroups(*awselbv2.DescribeTargetGroupsInput) (*awselbv2.DescribeTargetGroupsOutput, error)
	DeleteTargetGroup(*awselbv2.DeleteTargetGroupInput) (*awselbv2.DeleteTargetGroupOutput, error)
}

type TargetGroups struct {
	client targetGroupsClient
	logger logger
}

func NewTargetGroups(client targetGroupsClient, logger logger) TargetGroups {
	return TargetGroups{
		client: client,
		logger: logger,
	}
}

func (t TargetGroups) List(filter string) ([]common.Deletable, error) {
	targetGroups, err := t.client.DescribeTargetGroups(&awselbv2.DescribeTargetGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing target groups: %s", err)
	}

	var resources []common.Deletable
	for _, g := range targetGroups.TargetGroups {
		resource := NewTargetGroup(t.client, g.TargetGroupName, g.TargetGroupArn)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := t.logger.Prompt(fmt.Sprintf("Are you sure you want to delete target group %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
