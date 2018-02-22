package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Outputs struct {
	logger           logger
	terraformManager terraformManager
	stateValidator   stateValidator
}

func NewOutputs(logger logger, terraformManager terraformManager, stateValidator stateValidator) Outputs {
	return Outputs{
		logger:           logger,
		terraformManager: terraformManager,
		stateValidator:   stateValidator,
	}
}

func (o Outputs) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := o.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (o Outputs) Execute(subcommandFlags []string, state storage.State) error {
	outputs, err := o.terraformManager.GetOutputs()
	if err != nil {
		return err
	}
	for k, v := range outputs.Map {
		o.logger.Printf("%s: %+v\n", k, v)
	}
	return nil
}
