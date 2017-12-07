package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AzureLBs struct {
	terraformManager terraformManager
	logger           logger
}

func NewAzureLBs(terraformManager terraformManager, logger logger) AzureLBs {
	return AzureLBs{
		terraformManager: terraformManager,
		logger:           logger,
	}
}

func (l AzureLBs) Execute(subcommandFlags []string, state storage.State) error {
	terraformOutputs, err := l.terraformManager.GetOutputs()
	if err != nil {
		return err
	}

	if state.LB.Type == "cf" {
		l.logger.Printf("CF LB: %s\n", terraformOutputs.GetString("application_gateway"))
	} else {
		return errors.New("no lbs found")
	}

	return nil
}
