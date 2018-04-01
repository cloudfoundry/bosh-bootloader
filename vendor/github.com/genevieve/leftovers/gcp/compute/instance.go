package compute

import (
	"fmt"
	"strings"

	gcpcompute "google.golang.org/api/compute/v1"
)

type Instance struct {
	client      instancesClient
	name        string
	clearerName string
	zone        string
}

func NewInstance(client instancesClient, name, zone string, tags *gcpcompute.Tags) Instance {
	clearerName := name

	extra := []string{}
	if tags != nil && len(tags.Items) > 0 {
		for _, tag := range tags.Items {
			extra = append(extra, tag)
		}
	}

	if len(extra) > 0 {
		clearerName = fmt.Sprintf("%s (%s)", name, strings.Join(extra, ", "))
	}

	return Instance{
		client:      client,
		name:        name,
		clearerName: clearerName,
		zone:        zone,
	}
}

func (i Instance) Delete() error {
	err := i.client.DeleteInstance(i.zone, i.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (i Instance) Name() string {
	return i.clearerName
}

func (i Instance) Type() string {
	return "Instance"
}
