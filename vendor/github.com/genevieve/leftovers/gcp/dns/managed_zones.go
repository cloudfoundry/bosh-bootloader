package dns

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
	gcpdns "google.golang.org/api/dns/v1"
)

type managedZonesClient interface {
	ListManagedZones() (*gcpdns.ManagedZonesListResponse, error)
	DeleteManagedZone(zone string) error
}

type recordSets interface {
	Delete(managedZone string) error
}

type ManagedZones struct {
	client     managedZonesClient
	recordSets recordSets
	logger     logger
}

func NewManagedZones(client managedZonesClient, recordSets recordSets, logger logger) ManagedZones {
	return ManagedZones{
		client:     client,
		recordSets: recordSets,
		logger:     logger,
	}
}

func (m ManagedZones) List(filter string) ([]common.Deletable, error) {
	managedZones, err := m.client.ListManagedZones()
	if err != nil {
		return nil, fmt.Errorf("Listing managed zones: %s", err)
	}

	var resources []common.Deletable
	for _, zone := range managedZones.ManagedZones {
		resource := NewManagedZone(m.client, m.recordSets, zone.Name)

		if !strings.Contains(resource.name, filter) {
			continue
		}

		proceed := m.logger.Prompt(fmt.Sprintf("Are you sure you want to delete managed zone %s?", resource.name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
