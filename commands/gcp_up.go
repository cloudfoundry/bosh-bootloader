package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPUp struct {
	stateStore     stateStore
	keyPairUpdater keyPairUpdater
	gcpProvider    gcpProvider
}

type GCPUpConfig struct {
	ServiceAccountKeyPath string
	ProjectID             string
	Zone                  string
	Region                string
}

type gcpKeyPairCreator interface {
	Create() (string, string, error)
}

type keyPairUpdater interface {
	Update(projectID string) (storage.KeyPair, error)
}

type gcpProvider interface {
	SetConfig(serviceAccountKey string) error
}

func NewGCPUp(stateStore stateStore, keyPairUpdater keyPairUpdater, gcpProvider gcpProvider) GCPUp {
	return GCPUp{
		stateStore:     stateStore,
		keyPairUpdater: keyPairUpdater,
		gcpProvider:    gcpProvider,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	if state.GCP.Empty() || !upConfig.empty() {
		gcpDetails, err := u.parseUpConfig(upConfig)
		if err != nil {
			return err
		}
		err = u.gcpProvider.SetConfig(gcpDetails.ServiceAccountKey)
		if err != nil {
			return err
		}

		if state.GCP.Empty() {
			keyPair, err := u.keyPairUpdater.Update(gcpDetails.ProjectID)
			if err != nil {
				return err
			}

			state.KeyPair = keyPair
		}

		state.IAAS = "gcp"
		state.GCP = gcpDetails

		if err := u.stateStore.Set(state); err != nil {
			return err
		}
	}

	return nil
}

func (u GCPUp) parseUpConfig(upConfig GCPUpConfig) (storage.GCP, error) {
	switch {
	case upConfig.ServiceAccountKeyPath == "":
		return storage.GCP{}, errors.New("GCP service account key must be provided")
	case upConfig.ProjectID == "":
		return storage.GCP{}, errors.New("GCP project ID must be provided")
	case upConfig.Region == "":
		return storage.GCP{}, errors.New("GCP region must be provided")
	case upConfig.Zone == "":
		return storage.GCP{}, errors.New("GCP zone must be provided")
	}

	sak, err := ioutil.ReadFile(upConfig.ServiceAccountKeyPath)
	if err != nil {
		return storage.GCP{}, fmt.Errorf("error reading service account key: %v", err)
	}

	var tmp interface{}
	err = json.Unmarshal(sak, &tmp)
	if err != nil {
		return storage.GCP{}, fmt.Errorf("error parsing service account key: %v", err)
	}

	return storage.GCP{
		ServiceAccountKey: string(sak),
		ProjectID:         upConfig.ProjectID,
		Zone:              upConfig.Zone,
		Region:            upConfig.Region,
	}, nil
}

func (c GCPUpConfig) empty() bool {
	return c.ServiceAccountKeyPath == "" && c.ProjectID == "" && c.Region == "" && c.Zone == ""
}
