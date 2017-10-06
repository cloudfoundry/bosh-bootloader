package commands

import (
	// "encoding/json"
	"errors"
	// "strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AzureLBs struct {
	terraformManager terraformOutputter
	logger           logger
}

func NewAzureLBs(terraformManager terraformOutputter, logger logger) AzureLBs {
	return AzureLBs{
		terraformManager: terraformManager,
		logger:           logger,
	}
}

func (l AzureLBs) Execute(subcommandFlags []string, state storage.State) error {
	// terraformOutputs, err := l.terraformManager.GetOutputs(state)
	// if err != nil {
	// 	return err
	// }
	
	l.logger.Printf("TODO niroy 123")
	
	return errors.New("no lbs found")
	// return nil
}
