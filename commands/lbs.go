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

func (l LBs) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := l.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (l LBs) Execute(subcommandFlags []string, state storage.State) error {
	switch state.IAAS {
	case "aws":
		if err := l.awsLBs.Execute(subcommandFlags, state); err != nil {
			return err
		}
	case "gcp":
		if err := l.gcpLBs.Execute(subcommandFlags, state); err != nil {
			return err
		}
	}

	return nil
}
