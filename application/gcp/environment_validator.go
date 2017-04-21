package gcp

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type EnvironmentValidator struct {
	boshClientProvider boshClientProvider
}

type boshClientProvider interface {
	Client(directorAddress, directorUsername, directorPassword string) bosh.Client
}

func NewEnvironmentValidator(boshClientProvider boshClientProvider) EnvironmentValidator {
	return EnvironmentValidator{
		boshClientProvider: boshClientProvider,
	}
}

func (e EnvironmentValidator) Validate(state storage.State) error {
	if !state.NoDirector {
		boshClient := e.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)
		_, err := boshClient.Info()
		if err != nil {
			return application.BBLNotFound

		}
	}

	return nil
}
