package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type instanceGroupsClient interface {
	ListInstanceGroups(zone string) (*gcpcompute.InstanceGroupList, error)
	DeleteInstanceGroup(zone, instanceGroup string) error
}

type InstanceGroups struct {
	client instanceGroupsClient
	logger logger
	zones  map[string]string
}

func NewInstanceGroups(client instanceGroupsClient, logger logger, zones map[string]string) InstanceGroups {
	return InstanceGroups{
		client: client,
		logger: logger,
		zones:  zones,
	}
}

func (i InstanceGroups) List(filter string) ([]common.Deletable, error) {
	groups := []*gcpcompute.InstanceGroup{}
	for _, zone := range i.zones {
		l, err := i.client.ListInstanceGroups(zone)
		if err != nil {
			return nil, fmt.Errorf("Listing instance groups for zone %s: %s", zone, err)
		}

		groups = append(groups, l.Items...)
	}

	var resources []common.Deletable
	for _, group := range groups {
		resource := NewInstanceGroup(i.client, group.Name, i.zones[group.Zone])

		if !strings.Contains(group.Name, filter) {
			continue
		}

		proceed := i.logger.Prompt(fmt.Sprintf("Are you sure you want to delete instance group %s?", group.Name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
