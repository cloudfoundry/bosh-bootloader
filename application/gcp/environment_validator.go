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
	Client(jumpbox storage.Jumpbox, directorAddress, directorUsername, directorPassword, directorCACert string) (bosh.Client, error)
}

func NewEnvironmentValidator(boshClientProvider boshClientProvider) EnvironmentValidator {
	return EnvironmentValidator{
		boshClientProvider: boshClientProvider,
	}
}

func (e EnvironmentValidator) Validate(state storage.State) error {
	if state.TFState == "" {
		return application.BBLNotFound
	}

	if !state.NoDirector {
		boshClient, err := e.boshClientProvider.Client(state.Jumpbox, state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword, state.BOSH.DirectorSSLCA)
		if err != nil {
			return err //not tested
		}
		_, err = boshClient.Info()
		if err != nil {
			return application.DirectorNotReachable
		}
	}

	return nil
}
