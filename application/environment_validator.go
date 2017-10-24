package application

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var BBLNotFound error = errors.New("A bbl environment could not be found, please create a new environment before running this command again.")
var DirectorNotReachable error = errors.New("We couldn't communicate to the director in your state file. You may need to run `bbl up`.")

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
	if !state.NoDirector {
		boshClient, err := e.boshClientProvider.Client(state.Jumpbox, state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword, state.BOSH.DirectorSSLCA)
		if err != nil {
			return fmt.Errorf("bosh client provider: %s", err)
		}
		_, err = boshClient.Info()
		if err != nil {
			return fmt.Errorf("%s %s", DirectorNotReachable, err)
		}
	}

	if state.IAAS == "gcp" && len(state.GCP.Zones) == 0 {
		return errors.New("bbl state is missing availability zones; have you run bbl up?")
	}

	return nil
}
