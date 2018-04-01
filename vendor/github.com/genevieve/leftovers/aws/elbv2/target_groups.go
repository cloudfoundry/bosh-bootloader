package elbv2

import (
	"fmt"
	"strings"

	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/genevieve/leftovers/aws/common"
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

func (t TargetGroups) ListOnly(filter string) ([]common.Deletable, error) {
	return t.get(filter)
}

func (t TargetGroups) List(filter string) ([]common.Deletable, error) {
	resources, err := t.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := t.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (t TargetGroups) get(filter string) ([]common.Deletable, error) {
	targetGroups, err := t.client.DescribeTargetGroups(&awselbv2.DescribeTargetGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describe ELBV2 Target Groups: %s", err)
	}

	var resources []common.Deletable
	for _, g := range targetGroups.TargetGroups {
		resource := NewTargetGroup(t.client, g.TargetGroupName, g.TargetGroupArn)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
