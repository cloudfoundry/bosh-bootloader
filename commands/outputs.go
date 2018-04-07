package commands

import (
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type carto interface {
	Ymlize(string) (string, error)
}

type Outputs struct {
	logger         logger
	carto          carto
	stateStore     stateStore
	stateValidator stateValidator
}

func NewOutputs(logger logger, carto carto, stateStore stateStore, stateValidator stateValidator) Outputs {
	return Outputs{
		logger:         logger,
		carto:          carto,
		stateStore:     stateStore,
		stateValidator: stateValidator,
	}
}

func (o Outputs) CheckFastFails(subcommandFlags []string, state storage.State) error {
	return o.stateValidator.Validate()
}

func (o Outputs) Execute(subcommandFlags []string, state storage.State) error {
	dir, err := o.stateStore.GetVarsDir()
	if err != nil {
		return err
	}

	tfstate := filepath.Join(dir, terraform.TFSTATE)

	yml, err := o.carto.Ymlize(tfstate)
	if err != nil {
		return err
	}

	o.logger.Printf(string(yml))
	return nil
}
