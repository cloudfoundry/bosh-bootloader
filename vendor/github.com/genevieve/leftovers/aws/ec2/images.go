package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/genevieve/leftovers/common"
)

type imagesClient interface {
	DescribeImages(*awsec2.DescribeImagesInput) (*awsec2.DescribeImagesOutput, error)
	DeregisterImage(*awsec2.DeregisterImageInput) (*awsec2.DeregisterImageOutput, error)
}

type stsClient interface {
	GetCallerIdentity(*awssts.GetCallerIdentityInput) (*awssts.GetCallerIdentityOutput, error)
}

type Images struct {
	client       imagesClient
	stsClient    stsClient
	logger       logger
	resourceTags resourceTags
}

func NewImages(client imagesClient, stsClient stsClient, logger logger, resourceTags resourceTags) Images {
	return Images{
		client:       client,
		stsClient:    stsClient,
		logger:       logger,
		resourceTags: resourceTags,
	}
}

func (i Images) List(filter string) ([]common.Deletable, error) {
	caller, err := i.stsClient.GetCallerIdentity(&awssts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("Get caller identity: %s", err)
	}

	images, err := i.client.DescribeImages(&awsec2.DescribeImagesInput{
		Owners: []*string{caller.Account},
	})
	if err != nil {
		return nil, fmt.Errorf("Describing EC2 Images: %s", err)
	}

	var resources []common.Deletable
	for _, image := range images.Images {
		r := NewImage(i.client, image.ImageId, i.resourceTags)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := i.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (i Images) Type() string {
	return "ec2-image"
}
