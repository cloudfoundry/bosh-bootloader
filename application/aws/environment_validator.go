package aws

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type EnvironmentValidator struct {
	infrastructureManager infrastructureManager
	boshClientProvider    boshClientProvider
}

type infrastructureManager interface {
	Exists(stackName string) (bool, error)
}

type boshClientProvider interface {
	Client(directorAddress, directorUsername, directorPassword string) bosh.Client
}

func NewEnvironmentValidator(infrastructureManager infrastructureManager, boshClientProvider boshClientProvider) EnvironmentValidator {
	return EnvironmentValidator{
		infrastructureManager: infrastructureManager,
		boshClientProvider:    boshClientProvider,
	}
}

func (e EnvironmentValidator) Validate(state storage.State) error {
	if state.Stack.Name == "" && state.TFState == "" {
		return application.BBLNotFound
	}

	if state.Stack.Name != "" {
		stackExists, err := e.infrastructureManager.Exists(state.Stack.Name)
		if err != nil {
			return err
		}

		if !stackExists {
			return application.BBLNotFound
		}
	}

	if !state.NoDirector {
		boshClient := e.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)
		_, err := boshClient.Info()
		if err != nil {
			return application.BBLNotFound
		}
	}

	return nil
}
