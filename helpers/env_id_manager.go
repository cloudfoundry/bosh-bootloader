package helpers

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var matchString = regexp.MatchString

type EnvIDManager struct {
	envIDGenerator envIDGenerator
	networkClient  NetworkClient
}

type envIDGenerator interface {
	Generate() (string, error)
}

type NetworkClient interface {
	CheckExists(networkName string) (bool, error)
}

func NewEnvIDManager(envIDGenerator envIDGenerator, networkClient NetworkClient) EnvIDManager {
	return EnvIDManager{
		envIDGenerator: envIDGenerator,
		networkClient:  networkClient,
	}
}

func (e EnvIDManager) Sync(state storage.State, envID string) (storage.State, error) {
	if state.EnvID != "" {
		return state, nil
	}

	err := e.checkFastFail(state.IAAS, envID)
	if err != nil {
		return storage.State{}, err
	}

	if envID == "" {
		state.EnvID, err = e.envIDGenerator.Generate()
		if err != nil {
			return storage.State{}, err
		}
	} else {
		err = e.validateName(envID)
		if err != nil {
			return storage.State{}, err
		}

		state.EnvID = envID
	}

	return state, nil
}

func (e EnvIDManager) checkFastFail(iaas, envID string) error {
	var networkName string

	switch iaas {
	case "aws":
		networkName = envID + "-vpc"
	case "azure":
		networkName = envID
	case "gcp":
		networkName = envID + "-network"
	case "vsphere":
		return nil
	case "openstack":
		return nil
	case "cloudstack":
		return nil
	}

	exists, err := e.networkClient.CheckExists(networkName)
	if err != nil {
		return err
	}

	if exists {
		return errors.New(fmt.Sprintf("It looks like a bbl environment already exists with the name '%s'. Please provide a different name.", envID))
	}

	return nil
}

func (e EnvIDManager) validateName(envID string) error {
	matched, err := matchString("^(?:[a-z](?:[-a-z0-9]*[a-z0-9])?)$", envID)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("Names must start with a letter, be all lowercase, and be alphanumeric or hyphenated.")
	}

	return nil
}
