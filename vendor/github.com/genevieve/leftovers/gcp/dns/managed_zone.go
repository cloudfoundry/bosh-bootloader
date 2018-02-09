package dns

import "fmt"

type ManagedZone struct {
	client     managedZonesClient
	recordSets recordSets
	name       string
}

func NewManagedZone(client managedZonesClient, recordSets recordSets, name string) ManagedZone {
	return ManagedZone{
		client:     client,
		recordSets: recordSets,
		name:       name,
	}
}

func (m ManagedZone) Delete() error {
	err := m.recordSets.Delete(m.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting record sets for zone %s: %s", m.name, err)
	}

	err = m.client.DeleteManagedZone(m.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting managed zone %s: %s", m.name, err)
	}

	return nil
}

func (m ManagedZone) Name() string {
	return m.name
}
