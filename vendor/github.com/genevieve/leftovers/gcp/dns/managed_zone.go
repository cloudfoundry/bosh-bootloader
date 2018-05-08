package dns

import "fmt"

type ManagedZone struct {
	client     managedZonesClient
	recordSets recordSets
	name       string
	kind       string
}

func NewManagedZone(client managedZonesClient, recordSets recordSets, name string) ManagedZone {
	return ManagedZone{
		client:     client,
		recordSets: recordSets,
		name:       name,
		kind:       "managed-zone",
	}
}

func (m ManagedZone) Delete() error {
	err := m.recordSets.Delete(m.name)

	if err != nil {
		return fmt.Errorf("Delete record sets: %s", err)
	}

	err = m.client.DeleteManagedZone(m.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (m ManagedZone) Name() string {
	return m.name
}

func (m ManagedZone) Type() string {
	return "DNS Managed Zone"
}

func (m ManagedZone) Kind() string {
	return m.kind
}
