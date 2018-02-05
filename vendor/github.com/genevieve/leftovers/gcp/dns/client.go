package dns

import (
	gcpdns "google.golang.org/api/dns/v1"
)

type client struct {
	project string
	logger  logger

	managedZones *gcpdns.ManagedZonesService
	changes      *gcpdns.ChangesService
	recordSets   *gcpdns.ResourceRecordSetsService
}

func NewClient(project string, service *gcpdns.Service, logger logger) client {
	return client{
		project:      project,
		logger:       logger,
		managedZones: service.ManagedZones,
		changes:      service.Changes,
		recordSets:   service.ResourceRecordSets,
	}
}

func (c client) ListManagedZones() (*gcpdns.ManagedZonesListResponse, error) {
	return c.managedZones.List(c.project).Do()
}

func (c client) DeleteManagedZone(managedZone string) error {
	return c.managedZones.Delete(c.project, managedZone).Do()
}

func (c client) ListRecordSets(managedZone string) (*gcpdns.ResourceRecordSetsListResponse, error) {
	return c.recordSets.List(c.project, managedZone).Do()
}

func (c client) DeleteRecordSets(managedZone string, change *gcpdns.Change) error {
	_, err := c.changes.Create(c.project, managedZone, change).Do()
	return err
}
