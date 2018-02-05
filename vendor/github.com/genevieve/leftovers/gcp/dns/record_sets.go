package dns

import (
	"fmt"

	gcpdns "google.golang.org/api/dns/v1"
)

type recordSetsClient interface {
	ListRecordSets(managedZone string) (*gcpdns.ResourceRecordSetsListResponse, error)
	DeleteRecordSets(managedZone string, change *gcpdns.Change) error
}

type RecordSets struct {
	client recordSetsClient
}

func NewRecordSets(client recordSetsClient) RecordSets {
	return RecordSets{
		client: client,
	}
}

func (r RecordSets) Delete(managedZone string) error {
	recordSets, err := r.client.ListRecordSets(managedZone)
	if err != nil {
		return fmt.Errorf("Listing record sets: %s", err)
	}

	deletions := []*gcpdns.ResourceRecordSet{}
	for _, record := range recordSets.Rrsets {
		if record.Type == "NS" || record.Type == "SOA" {
			continue
		}

		deletions = append(deletions, record)
	}

	if len(deletions) > 0 {
		err = r.client.DeleteRecordSets(managedZone, &gcpdns.Change{
			Deletions: deletions,
		})
		if err != nil {
			return fmt.Errorf("Deleting record sets: %s", err)
		}
	}
	return nil
}
