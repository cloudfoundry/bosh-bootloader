package config

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPZonerHack struct {
	gcpAvailabilityZoneRetriever gcpAvailabilityZoneRetriever
}

type gcpAvailabilityZoneRetriever interface {
	GetZones(string) ([]string, error)
}

func NewGCPZonerHack(gcpAvailabilityZoneRetriever gcpAvailabilityZoneRetriever) GCPZonerHack {
	return GCPZonerHack{
		gcpAvailabilityZoneRetriever: gcpAvailabilityZoneRetriever,
	}
}

func (g GCPZonerHack) SetZones(state storage.State) (storage.State, error) {
	if len(state.GCP.Zones) == 0 {
		zones, err := g.gcpAvailabilityZoneRetriever.GetZones(state.GCP.Region)
		if err != nil {
			return storage.State{}, fmt.Errorf("Retrieving availability zones: %s", err)
		}
		if len(zones) == 0 {
			return storage.State{}, errors.New("Zone list is empty")
		}
		state.GCP.Zones = zones
	}

	if state.GCP.Zone == "" {
		state.GCP.Zone = state.GCP.Zones[0]
	}

	return state, nil
}
