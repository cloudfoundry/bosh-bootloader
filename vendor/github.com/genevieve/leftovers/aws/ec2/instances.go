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
	client       instancesClient
	logger       logger
	resourceTags resourceTags
}

func NewInstances(client instancesClient, logger logger, resourceTags resourceTags) Instances {
	return Instances{
		client:       client,
		logger:       logger,
		resourceTags: resourceTags,
	}
}

func (a Instances) ListOnly(filter string) ([]common.Deletable, error) {
	return a.get(filter)
}

func (a Instances) List(filter string) ([]common.Deletable, error) {
	resources, err := a.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := a.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (a Instances) get(filter string) ([]common.Deletable, error) {
	instances, err := a.client.DescribeInstances(&awsec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing EC2 Instances: %s", err)
	}

	var resources []common.Deletable
	for _, r := range instances.Reservations {
		for _, i := range r.Instances {
			resource := NewInstance(a.client, a.resourceTags, i.InstanceId, i.KeyName, i.Tags)

			if a.alreadyShutdown(*i.State.Name) {
				continue
			}

			if !strings.Contains(resource.identifier, filter) {
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
