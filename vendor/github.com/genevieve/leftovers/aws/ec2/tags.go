package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/common"
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

func (a Tags) List(filter string) ([]common.Deletable, error) {
	output, err := a.client.DescribeTags(&awsec2.DescribeTagsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describe EC2 Tags: %s", err)
	}

	var resources []common.Deletable
	for _, t := range output.Tags {
		if *t.ResourceId != "" {
			continue
		}

		r := NewTag(a.client, t.Key, t.Value, t.ResourceId)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := a.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (a Tags) Type() string {
	return "ec2-tag"
}
