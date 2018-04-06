package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Image struct {
	client       imagesClient
	id           *string
	identifier   string
	resourceTags resourceTags
}

func NewImage(client imagesClient, id *string, resourceTags resourceTags) Image {
	return Image{
		client:       client,
		id:           id,
		identifier:   *id,
		resourceTags: resourceTags,
	}
}

func (i Image) Delete() error {
	_, err := i.client.DeregisterImage(&awsec2.DeregisterImageInput{ImageId: i.id})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	err = i.resourceTags.Delete("image", *i.id)
	if err != nil {
		return fmt.Errorf("Delete tags: %s", err)
	}

	return nil
}

func (i Image) Name() string {
	return i.identifier
}

func (i Image) Type() string {
	return "EC2 Image"
}
