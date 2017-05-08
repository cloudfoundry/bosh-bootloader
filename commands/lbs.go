package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	LBsCommand = "lbs"
)

type LBs struct {
	gcpLBs         gcpLBs
	awsLBs         awsLBs
	stateValidator stateValidator
	logger         logger
}

type gcpLBs interface {
	Execute([]string, storage.State) error
}

type awsLBs interface {
	Execute([]string, storage.State) error
}

func NewLBs(gcpLBs gcpLBs, awsLBs awsLBs, stateValidator stateValidator, logger logger) LBs {
	return LBs{
		gcpLBs:         gcpLBs,
		awsLBs:         awsLBs,
		stateValidator: stateValidator,
		logger:         logger,
	}
}

func (c LBs) Execute(subcommandFlags []string, state storage.State) error {
	err := c.stateValidator.Validate()
	if err != nil {
		return err
	}

	switch state.IAAS {
	case "aws":
		if err := c.awsLBs.Execute(subcommandFlags, state); err != nil {
			return err
		}
	case "gcp":
		if err := c.gcpLBs.Execute(subcommandFlags, state); err != nil {
			return err
		}
	}

	return nil
}
