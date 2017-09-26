package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPUp struct {
	gcpAvailabilityZoneRetriever gcpAvailabilityZoneRetriever
}

type gcpAvailabilityZoneRetriever interface {
	GetZones(string) ([]string, error)
}

func NewGCPUp(gcpAvailabilityZoneRetriever gcpAvailabilityZoneRetriever) GCPUp {
	return GCPUp{
		gcpAvailabilityZoneRetriever: gcpAvailabilityZoneRetriever,
	}
}

func (u GCPUp) Execute(state storage.State) (storage.State, error) {
	var err error
	state.GCP.Zones, err = u.gcpAvailabilityZoneRetriever.GetZones(state.GCP.Region)
	if err != nil {
		return storage.State{}, fmt.Errorf("Retrieving availability zones: %s", err)
	}

	return state, nil
}
