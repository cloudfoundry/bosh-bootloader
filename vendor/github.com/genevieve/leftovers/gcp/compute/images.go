package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type imagesClient interface {
	ListImages() (*gcpcompute.ImageList, error)
	DeleteImage(image string) error
}

type Images struct {
	client imagesClient
	logger logger
}

func NewImages(client imagesClient, logger logger) Images {
	return Images{
		client: client,
		logger: logger,
	}
}

func (d Images) List(filter string) ([]common.Deletable, error) {
	images, err := d.client.ListImages()
	if err != nil {
		return nil, fmt.Errorf("Listing images: %s", err)
	}

	var resources []common.Deletable
	for _, image := range images.Items {
		resource := NewImage(d.client, image.Name)

		if !strings.Contains(image.Name, filter) {
			continue
		}

		proceed := d.logger.Prompt(fmt.Sprintf("Are you sure you want to delete image %s?", image.Name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
