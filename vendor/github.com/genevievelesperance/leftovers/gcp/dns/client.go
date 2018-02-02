package dns

import (
	gcpdns "google.golang.org/api/dns/v1"
)

type client struct {
	project string
	logger  logger

	managedZones *gcpdns.ManagedZonesService
}

func NewClient(project string, service *gcpdns.Service, logger logger) client {
	return client{
		project:      project,
		logger:       logger,
		managedZones: service.ManagedZones,
	}
}

func (c client) ListManagedZones() (*gcpdns.ManagedZonesListResponse, error) {
	return c.managedZones.List(c.project).Do()
}

func (c client) DeleteManagedZone(managedZone string) error {
	return c.managedZones.Delete(c.project, managedZone).Do()
}
