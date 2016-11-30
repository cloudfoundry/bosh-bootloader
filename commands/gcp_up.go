package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var (
	writeFile = ioutil.WriteFile
	tempDir   = ioutil.TempDir
)

type GCPUp struct {
	stateStore       stateStore
	keyPairUpdater   keyPairUpdater
	gcpProvider      gcpProvider
	terraformApplier terraformApplier
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

type terraformApplier interface {
	Apply(credentials, envID, projectID, zone, region, template, tfState string) (string, error)
}

func NewGCPUp(stateStore stateStore, keyPairUpdater keyPairUpdater, gcpProvider gcpProvider, terraformApplier terraformApplier) GCPUp {
	return GCPUp{
		stateStore:       stateStore,
		keyPairUpdater:   keyPairUpdater,
		gcpProvider:      gcpProvider,
		terraformApplier: terraformApplier,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	if !upConfig.empty() {
		gcpDetails, err := u.parseUpConfig(upConfig)
		if err != nil {
			return err
		}

		state.IAAS = "gcp"
		state.GCP = gcpDetails
	}

	if err := u.validateState(state); err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	if err := u.gcpProvider.SetConfig(state.GCP.ServiceAccountKey); err != nil {
		return err
	}

	if state.KeyPair.IsEmpty() {
		keyPair, err := u.keyPairUpdater.Update(state.GCP.ProjectID)
		if err != nil {
			return err
		}
		state.KeyPair = keyPair
		if err := u.stateStore.Set(state); err != nil {
			return err
		}
	}

	tempDir, err := tempDir("", "")
	if err != nil {
		panic(err)
	}

	serviceAccountKeyPath := filepath.Join(tempDir, "credentials.json")
	err = writeFile(serviceAccountKeyPath, []byte(state.GCP.ServiceAccountKey), os.ModePerm)
	if err != nil {
		panic(err)
	}

	tfState, err := u.terraformApplier.Apply(serviceAccountKeyPath, state.EnvID, state.GCP.ProjectID, state.GCP.Zone, state.GCP.Region, terraformTemplate, state.TFState)
	if err != nil {
		return err
	}

	state.TFState = tfState
	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	return nil
}

func (u GCPUp) validateState(state storage.State) error {
	switch {
	case state.GCP.ServiceAccountKey == "":
		return errors.New("GCP service account key must be provided")
	case state.GCP.ProjectID == "":
		return errors.New("GCP project ID must be provided")
	case state.GCP.Region == "":
		return errors.New("GCP region must be provided")
	case state.GCP.Zone == "":
		return errors.New("GCP zone must be provided")
	}

	return nil
}

func (u GCPUp) parseUpConfig(upConfig GCPUpConfig) (storage.GCP, error) {
	if upConfig.ServiceAccountKeyPath == "" {
		return storage.GCP{}, errors.New("GCP service account key must be provided")
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
