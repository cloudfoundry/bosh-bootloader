package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/common"
)

type instancesClient interface {
	DescribeInstances(*awsec2.DescribeInstancesInput) (*awsec2.DescribeInstancesOutput, error)
	TerminateInstances(*awsec2.TerminateInstancesInput) (*awsec2.TerminateInstancesOutput, error)

	DescribeAddresses(*awsec2.DescribeAddressesInput) (*awsec2.DescribeAddressesOutput, error)
	ReleaseAddress(*awsec2.ReleaseAddressInput) (*awsec2.ReleaseAddressOutput, error)
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

func (i Instances) List(filter string) ([]common.Deletable, error) {
	instances, err := i.client.DescribeInstances(&awsec2.DescribeInstancesInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("pending"), aws.String("running"), aws.String("shutting-down"), aws.String("stopping"), aws.String("stopped")},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("Describing EC2 Instances: %s", err)
	}

	var resources []common.Deletable
	for _, r := range instances.Reservations {
		for _, instance := range r.Instances {
			r := NewInstance(i.client, i.logger, i.resourceTags, instance.InstanceId, instance.KeyName, instance.Tags)

			if !strings.Contains(r.Name(), filter) {
				continue
			}

			proceed := i.logger.PromptWithDetails(r.Type(), r.Name())
			if !proceed {
				continue
			}

			resources = append(resources, r)
		}
	}

	return resources, nil
}

func (i Instances) Type() string {
	return "ec2-instance"
}
