package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type instanceGroupManagersClient interface {
	ListInstanceGroupManagers(zone string) ([]*gcpcompute.InstanceGroupManager, error)
	DeleteInstanceGroupManager(zone, instanceGroupManager string) error
}

type InstanceGroupManagers struct {
	client instanceGroupManagersClient
	logger logger
	zones  map[string]string
}

func NewInstanceGroupManagers(client instanceGroupManagersClient, logger logger, zones map[string]string) InstanceGroupManagers {
	return InstanceGroupManagers{
		client: client,
		logger: logger,
		zones:  zones,
	}
}

func (i InstanceGroupManagers) List(filter string) ([]common.Deletable, error) {
	managers := []*gcpcompute.InstanceGroupManager{}
	for _, zone := range i.zones {
		l, err := i.client.ListInstanceGroupManagers(zone)
		if err != nil {
			return nil, fmt.Errorf("List Instance Group Managers for zone %s: %s", zone, err)
		}

		managers = append(managers, l...)
	}

	var resources []common.Deletable
	for _, manager := range managers {
		resource := NewInstanceGroupManager(i.client, manager.Name, i.zones[manager.Zone])

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

func (i InstanceGroupManagers) Type() string {
	return "instance-group-manager"
}
