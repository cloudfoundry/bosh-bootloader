package helpers

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var matchString = regexp.MatchString

type EnvIDManager struct {
	envIDGenerator        envIDGenerator
	networkClient         NetworkClient
	infrastructureManager infrastructureManager
}

type envIDGenerator interface {
	Generate() (string, error)
}

type infrastructureManager interface {
	Exists(stackName string) (bool, error)
}

type NetworkClient interface {
	CheckExists(networkName string) (bool, error)
}

func NewEnvIDManager(envIDGenerator envIDGenerator, infrastructureManager infrastructureManager, networkClient NetworkClient) EnvIDManager {
	return EnvIDManager{
		envIDGenerator:        envIDGenerator,
		infrastructureManager: infrastructureManager,
		networkClient:         networkClient,
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

	err = e.validateName(envID)
	if err != nil {
		return storage.State{}, err
	}

	if envID != "" {
		state.EnvID = envID
	} else {
		state.EnvID, err = e.envIDGenerator.Generate()
		if err != nil {
			return storage.State{}, err
		}
	}

	return state, nil
}

func (e EnvIDManager) checkFastFail(iaas, envID string) error {
	if iaas == "aws" {
		stackName := "stack-" + envID
		stackExists, err := e.infrastructureManager.Exists(stackName)
		if err != nil {
			return err
		}
		if stackExists {
			return errors.New(fmt.Sprintf("It looks like a bbl environment already exists with the name '%s'. Please provide a different name.", envID))
		}
	}

	var networkName string
	switch iaas {
	case "gcp":
		networkName = envID + "-network"
	case "aws":
		networkName = envID + "-vpc"
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
	if envID == "" {
		return nil
	}

	matched, err := matchString("^(?:[a-z](?:[-a-z0-9]+[a-z0-9])?)$", envID)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("Names must start with a letter and be alphanumeric or hyphenated.")
	}

	return nil
}
