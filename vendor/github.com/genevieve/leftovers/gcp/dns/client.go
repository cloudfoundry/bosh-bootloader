package dns

import (
	gcpdns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
)

type client struct {
	project string

	managedZones *gcpdns.ManagedZonesService
	changes      *gcpdns.ChangesService
	recordSets   *gcpdns.ResourceRecordSetsService
}

func NewClient(project string, service *gcpdns.Service) client {
	return client{
		project:      project,
		managedZones: service.ManagedZones,
		changes:      service.Changes,
		recordSets:   service.ResourceRecordSets,
	}
}

func (c client) ListManagedZones() (*gcpdns.ManagedZonesListResponse, error) {
	return c.managedZones.List(c.project).Do()
}

func (c client) DeleteManagedZone(managedZone string) error {
	err := c.managedZones.Delete(c.project, managedZone).Do()

	return handleNotFoundError(err)
}

func (c client) ListRecordSets(managedZone string) (*gcpdns.ResourceRecordSetsListResponse, error) {
	return c.recordSets.List(c.project, managedZone).Do()
}

func (c client) DeleteRecordSets(managedZone string, change *gcpdns.Change) error {
	_, err := c.changes.Create(c.project, managedZone, change).Do()
	return handleNotFoundError(err)
}

func handleNotFoundError(err error) error {
	if err != nil {
		gerr, ok := err.(*googleapi.Error)
		if ok && gerr != nil && gerr.Code == 404 {
			return nil
		}

		return err
	}
	return nil
}
