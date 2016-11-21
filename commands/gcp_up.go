package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
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

	tfState, err := u.terraformApplier.Apply(upConfig.ServiceAccountKeyPath, state.EnvID, upConfig.ProjectID, upConfig.Zone, upConfig.Region, `variable "project_id" {
	type = "string"
}

variable "region" {
	type = "string"
}

variable "zone" {
	type = "string"
}

variable "env_id" {
	type = "string"
}

variable "credentials" {
	type = "string"
}

provider "google" {
	credentials = "${file("${var.credentials}")}"
	project = "${var.project_id}"
	region = "${var.region}"
}

resource "google_compute_network" "bbl" {
  name		 = "${var.env_id}"
}

resource "google_compute_subnetwork" "bbl-subnet" {
  name			= "${var.env_id}-subnet"
  ip_cidr_range = "10.0.0.0/16"
  network		= "${google_compute_network.bbl.self_link}"
}`, state.TFState)
	if err != nil {
		return err
	}

	state.TFState = tfState
	if err := u.stateStore.Set(state); err != nil {
		return err
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
