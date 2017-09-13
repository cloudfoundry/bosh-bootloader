package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type LBs struct {
	lbs            LBsCmd
	stateValidator stateValidator
}

type LBsCmd interface {
	Execute([]string, storage.State) error
}

func NewLBs(lbs LBsCmd, stateValidator stateValidator) LBs {
	return LBs{
		lbs:            lbs,
		stateValidator: stateValidator,
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
	return l.lbs.Execute(subcommandFlags, state)
}
