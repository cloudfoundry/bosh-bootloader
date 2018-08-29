package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type disksClient interface {
	ListDisks(zone string) ([]*gcpcompute.Disk, error)
	DeleteDisk(zone, disk string) error
}

type Disks struct {
	client disksClient
	logger logger
	zones  map[string]string
}

func NewDisks(client disksClient, logger logger, zones map[string]string) Disks {
	return Disks{
		client: client,
		logger: logger,
		zones:  zones,
	}
}

func (d Disks) List(filter string) ([]common.Deletable, error) {
	disks := []*gcpcompute.Disk{}
	for _, zone := range d.zones {
		l, err := d.client.ListDisks(zone)
		if err != nil {
			return nil, fmt.Errorf("List Disks for zone %s: %s", zone, err)
		}

		disks = append(disks, l...)
	}

	var resources []common.Deletable
	for _, disk := range disks {
		resource := NewDisk(d.client, disk.Name, d.zones[disk.Zone])

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		proceed := d.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (d Disks) Type() string {
	return "disk"
}
