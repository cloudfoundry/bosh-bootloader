package compute

import (
	"fmt"
	"strings"
	"sync"

	gcpcompute "google.golang.org/api/compute/v1"
)

type instancesClient interface {
	ListInstances(zone string) (*gcpcompute.InstanceList, error)
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

func (i Instances) List(filter string) (map[string]string, error) {
	instances := []*gcpcompute.Instance{}
	for _, zone := range i.zones {
		l, err := i.client.ListInstances(zone)
		if err != nil {
			return nil, fmt.Errorf("Listing instances for zone %s: %s", zone, err)
		}

		instances = append(instances, l.Items...)
	}

	delete := map[string]string{}
	for _, instance := range instances {
		n := i.clearerName(instance)

		if !strings.Contains(n, filter) {
			continue
		}

		proceed := i.logger.Prompt(fmt.Sprintf("Are you sure you want to delete instance %s?", n))
		if !proceed {
			continue
		}

		delete[instance.Name] = i.zones[instance.Zone]
	}

	return delete, nil
}

func (i Instances) Delete(instances map[string]string) {
	var wg sync.WaitGroup

	for name, zone := range instances {
		wg.Add(1)

		go func(name, zone string) {
			err := i.client.DeleteInstance(zone, name)

			if err != nil {
				i.logger.Printf("ERROR deleting instance %s: %s\n", name, err)
			} else {
				i.logger.Printf("SUCCESS deleting instance %s\n", name)
			}

			wg.Done()
		}(name, zone)
	}

	wg.Wait()
}

func (s Instances) clearerName(i *gcpcompute.Instance) string {
	extra := []string{}
	if i.Tags != nil && len(i.Tags.Items) > 0 {
		for _, tag := range i.Tags.Items {
			extra = append(extra, tag)
		}
	}

	if len(extra) > 0 {
		return fmt.Sprintf("%s (%s)", i.Name, strings.Join(extra, ", "))
	}

	return i.Name
}
