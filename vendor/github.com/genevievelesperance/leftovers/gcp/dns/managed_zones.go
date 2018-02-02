package dns

import (
	"fmt"
	"strings"
	"sync"

	gcpdns "google.golang.org/api/dns/v1"
)

type managedZonesClient interface {
	ListManagedZones() (*gcpdns.ManagedZonesListResponse, error)
	DeleteManagedZone(zone string) error
}

type ManagedZones struct {
	client managedZonesClient
	logger logger
}

func NewManagedZones(client managedZonesClient, logger logger) ManagedZones {
	return ManagedZones{
		client: client,
		logger: logger,
	}
}

func (m ManagedZones) List(filter string) (map[string]string, error) {
	managedZones, err := m.client.ListManagedZones()
	if err != nil {
		return nil, fmt.Errorf("Listing managed zones: %s", err)
	}

	delete := map[string]string{}
	for _, zone := range managedZones.ManagedZones {
		if !strings.Contains(zone.Name, filter) {
			continue
		}

		proceed := m.logger.Prompt(fmt.Sprintf("Are you sure you want to delete managed zone %s?", zone.Name))
		if !proceed {
			continue
		}

		delete[zone.Name] = ""
	}

	return delete, nil
}

func (m ManagedZones) Delete(zones map[string]string) {
	var wg sync.WaitGroup

	for name, _ := range zones {
		wg.Add(1)

		go func(name string) {
			err := m.client.DeleteManagedZone(name)

			if err != nil {
				m.logger.Printf("ERROR deleting managed zone %s: %s\n", name, err)
			} else {
				m.logger.Printf("SUCCESS deleting managed zone %s\n", name)
			}
			wg.Done()
		}(name)
	}

	wg.Wait()
}
