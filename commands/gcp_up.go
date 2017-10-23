package commands

import (
	"errors"
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
	zones, err := u.gcpAvailabilityZoneRetriever.GetZones(state.GCP.Region)
	if err != nil {
		return storage.State{}, fmt.Errorf("Retrieving availability zones: %s", err)
	}
	if len(zones) == 0 {
		return storage.State{}, errors.New("Zone list is empty")
	}

	state.GCP.Zones = zones
	if state.GCP.Zone == "" {
		state.GCP.Zone = zones[0]
	}

	return state, nil
}
