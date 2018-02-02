package compute

import (
	"fmt"
	"strings"
	"sync"

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

func (i InstanceGroups) List(filter string) (map[string]string, error) {
	groups := []*gcpcompute.InstanceGroup{}
	for _, zone := range i.zones {
		l, err := i.client.ListInstanceGroups(zone)
		if err != nil {
			return nil, fmt.Errorf("Listing instance groups for zone %s: %s", zone, err)
		}

		groups = append(groups, l.Items...)
	}

	delete := map[string]string{}
	for _, group := range groups {
		if !strings.Contains(group.Name, filter) {
			continue
		}

		proceed := i.logger.Prompt(fmt.Sprintf("Are you sure you want to delete instance group %s?", group.Name))
		if !proceed {
			continue
		}

		delete[group.Name] = i.zones[group.Zone]
	}

	return delete, nil
}

func (i InstanceGroups) Delete(groups map[string]string) {
	var wg sync.WaitGroup

	for name, zone := range groups {
		wg.Add(1)

		go func(name, zone string) {
			err := i.client.DeleteInstanceGroup(zone, name)

			if err != nil {
				i.logger.Printf("ERROR deleting instance group %s: %s\n", name, err)
			} else {
				i.logger.Printf("SUCCESS deleting instance group %s\n", name)
			}

			wg.Done()
		}(name, zone)
	}

	wg.Wait()
}
