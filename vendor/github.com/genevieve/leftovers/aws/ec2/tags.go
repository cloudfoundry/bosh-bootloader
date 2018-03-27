package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
)

type tagsClient interface {
	DescribeTags(*awsec2.DescribeTagsInput) (*awsec2.DescribeTagsOutput, error)
	DeleteTags(*awsec2.DeleteTagsInput) (*awsec2.DeleteTagsOutput, error)
}

type Tags struct {
	client tagsClient
	logger logger
}

func NewTags(client tagsClient, logger logger) Tags {
	return Tags{
		client: client,
		logger: logger,
	}
}

func (a Tags) ListAll(filter string) ([]common.Deletable, error) {
	return a.get(filter)
}

func (a Tags) List(filter string) ([]common.Deletable, error) {
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

func (a Tags) get(filter string) ([]common.Deletable, error) {
	output, err := a.client.DescribeTags(&awsec2.DescribeTagsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing EC2 Tags: %s", err)
	}

	var resources []common.Deletable
	for _, t := range output.Tags {
		resource := NewTag(a.client, t.Key, t.Value, t.ResourceId)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
