package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type imagesClient interface {
	ListImages() ([]*gcpcompute.Image, error)
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

func (i Images) List(filter string) ([]common.Deletable, error) {
	images, err := i.client.ListImages()
	if err != nil {
		return nil, fmt.Errorf("List Images: %s", err)
	}

	var resources []common.Deletable
	for _, image := range images {
		resource := NewImage(i.client, image.Name)

		if !strings.Contains(image.Name, filter) {
			continue
		}

		proceed := i.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (i Images) Type() string {
	return "image"
}
