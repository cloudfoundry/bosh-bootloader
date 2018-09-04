package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type instanceTemplatesClient interface {
	ListInstanceTemplates() ([]*gcpcompute.InstanceTemplate, error)
	DeleteInstanceTemplate(template string) error
}

type InstanceTemplates struct {
	client instanceTemplatesClient
	logger logger
}

func NewInstanceTemplates(client instanceTemplatesClient, logger logger) InstanceTemplates {
	return InstanceTemplates{
		client: client,
		logger: logger,
	}
}

func (i InstanceTemplates) List(filter string) ([]common.Deletable, error) {
	templates, err := i.client.ListInstanceTemplates()
	if err != nil {
		return nil, fmt.Errorf("List Instance Templates: %s", err)
	}

	var resources []common.Deletable
	for _, template := range templates {
		resource := NewInstanceTemplate(i.client, template.Name)

		if !strings.Contains(resource.Name(), filter) {
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

func (i InstanceTemplates) Type() string {
	return "instance-template"
}
