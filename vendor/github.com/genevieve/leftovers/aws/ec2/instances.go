package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
)

type instancesClient interface {
	DescribeInstances(*awsec2.DescribeInstancesInput) (*awsec2.DescribeInstancesOutput, error)
	TerminateInstances(*awsec2.TerminateInstancesInput) (*awsec2.TerminateInstancesOutput, error)
}

type Instances struct {
	client instancesClient
	logger logger
}

func NewInstances(client instancesClient, logger logger) Instances {
	return Instances{
		client: client,
		logger: logger,
	}
}

func (a Instances) List(filter string) ([]common.Deletable, error) {
	instances, err := a.client.DescribeInstances(&awsec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing instances: %s", err)
	}

	var resources []common.Deletable
	for _, r := range instances.Reservations {
		for _, i := range r.Instances {
			resource := NewInstance(a.client, i.InstanceId, i.KeyName, i.Tags)

			if a.alreadyShutdown(*i.State.Name) {
				continue
			}

			if !strings.Contains(resource.identifier, filter) {
				continue
			}

			proceed := a.logger.Prompt(fmt.Sprintf("Are you sure you want to terminate instance %s?", resource.identifier))
			if !proceed {
				continue
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

func (a Instances) alreadyShutdown(state string) bool {
	return state == "shutting-down" || state == "terminated"
}
