package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPUp struct {
	stateStore stateStore
}

type GCPUpConfig struct {
	ServiceAccountKeyPath string
	ProjectID             string
	Zone                  string
	Region                string
}

func NewGCPUp(stateStore stateStore) GCPUp {
	return GCPUp{
		stateStore: stateStore,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	state.IAAS = "gcp"

	if state.GCP.Empty() || !upConfig.empty() {
		err := u.validateUpConfig(upConfig)
		if err != nil {
			return err
		}

		state.GCP = storage.GCP{
			ServiceAccountKey: upConfig.ServiceAccountKeyPath,
			ProjectID:         upConfig.ProjectID,
			Zone:              upConfig.Zone,
			Region:            upConfig.Region,
		}
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	return nil
}

func (u GCPUp) validateUpConfig(upConfig GCPUpConfig) error {
	switch {
	case upConfig.ServiceAccountKeyPath == "":
		return errors.New("GCP service account key must be provided")
	case upConfig.ProjectID == "":
		return errors.New("GCP project ID must be provided")
	case upConfig.Region == "":
		return errors.New("GCP region must be provided")
	case upConfig.Zone == "":
		return errors.New("GCP zone must be provided")
	}

	return nil
}

func (c GCPUpConfig) empty() bool {
	return c.ServiceAccountKeyPath == "" && c.ProjectID == "" && c.Region == "" && c.Zone == ""
}
