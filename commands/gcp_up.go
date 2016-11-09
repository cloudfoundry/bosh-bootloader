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
	keyPairCreator gcpKeyPairCreator
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

func NewGCPUp(stateStore stateStore, keyPairCreator gcpKeyPairCreator) GCPUp {
	return GCPUp{
		stateStore:     stateStore,
		keyPairCreator: keyPairCreator,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	state.IAAS = "gcp"

	if state.GCP.Empty() || !upConfig.empty() {
		err := u.validateUpConfig(upConfig)
		if err != nil {
			return err
		}

		sak, err := ioutil.ReadFile(upConfig.ServiceAccountKeyPath)
		if err != nil {
			return fmt.Errorf("error reading service account key: %v", err)
		}

		var tmp interface{}
		err = json.Unmarshal(sak, &tmp)
		if err != nil {
			return fmt.Errorf("error parsing service account key: %v", err)
		}

		state.GCP = storage.GCP{
			ServiceAccountKey: string(sak),
			ProjectID:         upConfig.ProjectID,
			Zone:              upConfig.Zone,
			Region:            upConfig.Region,
		}

		privateKey, publicKey, err := u.keyPairCreator.Create()
		if err != nil {
			return err
		}

		state.KeyPair.PrivateKey = privateKey
		state.KeyPair.PublicKey = publicKey
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
