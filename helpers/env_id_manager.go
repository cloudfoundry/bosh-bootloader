package helpers

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type EnvIDManager struct {
	envIDGenerator        envIDGenerator
	gcpClientProvider     gcpClientProvider
	infrastructureManager infrastructureManager
}

type envIDGenerator interface {
	Generate() (string, error)
}

type infrastructureManager interface {
	Exists(stackName string) (bool, error)
}

type gcpClientProvider interface {
	Client() gcp.Client
}

func NewEnvIDManager(envIDGenerator envIDGenerator, gcpClientProvider gcpClientProvider,
	infrastructureManager infrastructureManager) EnvIDManager {
	return EnvIDManager{
		envIDGenerator:        envIDGenerator,
		gcpClientProvider:     gcpClientProvider,
		infrastructureManager: infrastructureManager,
	}
}

func (e EnvIDManager) Sync(state storage.State, envID string) (string, error) {
	if state.EnvID != "" {
		return state.EnvID, nil
	}

	err := e.checkFastFail(state.IAAS, envID)
	if err != nil {
		return "", err
	}

	if envID != "" {
		return envID, nil
	}

	return e.envIDGenerator.Generate()
}

func (e EnvIDManager) checkFastFail(iaas, envID string) error {
	switch iaas {
	case "gcp":
		gcpClient := e.gcpClientProvider.Client()
		networkName := envID + "-network"
		networkList, err := gcpClient.GetNetworks(networkName)
		if err != nil {
			return err
		}
		if len(networkList.Items) > 0 {
			return errors.New(fmt.Sprintf("It looks like a bbl environment already exists with the name '%s'. Please provide a different name.", envID))
		}
	case "aws":
		stackName := "stack-" + envID
		stackExists, err := e.infrastructureManager.Exists(stackName)
		if err != nil {
			return err
		}
		if stackExists {
			return errors.New(fmt.Sprintf("It looks like a bbl environment already exists with the name '%s'. Please provide a different name.", envID))
		}
	}
	return nil
}
