package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type instancesClient interface {
	ListInstances(zone string) ([]*gcpcompute.Instance, error)
	DeleteInstance(zone, instance string) error
}

type Instances struct {
	client instancesClient
	logger logger
	zones  map[string]string
}

func NewInstances(client instancesClient, logger logger, zones map[string]string) Instances {
	return Instances{
		client: client,
		logger: logger,
		zones:  zones,
	}
}

func (i Instances) List(filter string) ([]common.Deletable, error) {
	instances := []*gcpcompute.Instance{}
	for _, zone := range i.zones {
		l, err := i.client.ListInstances(zone)
		if err != nil {
			return nil, fmt.Errorf("List Instances for zone %s: %s", zone, err)
		}

		instances = append(instances, l...)
	}

	var resources []common.Deletable
	for _, instance := range instances {
		resource := NewInstance(i.client, instance.Name, i.zones[instance.Zone], instance.Tags)

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

func (i Instances) Type() string {
	return "compute-instance"
}
