package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	yaml "gopkg.in/yaml.v2"
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
	return o.stateValidator.Validate()
}

func (o Outputs) Execute(subcommandFlags []string, state storage.State) error {
	outputs, err := o.terraformManager.GetOutputs()
	if err != nil {
		return err
	}
	marshalled, err := yaml.Marshal(outputs.Map)
	if err != nil {
		return err
	}
	o.logger.Printf(string(marshalled))
	return nil
}
