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

	switch state.LB.Type {
	case "cf":
		l.logger.Printf("CF LB: %s\n", terraformOutputs.GetString("cf_app_gateway_name"))
	case "concourse":
		l.logger.Printf("Concourse LB: %s (%s)\n", terraformOutputs.GetString("concourse_lb_name"), terraformOutputs.GetString("concourse_lb_ip"))
	default:
		return errors.New("no lbs found")
	}

	return nil
}
