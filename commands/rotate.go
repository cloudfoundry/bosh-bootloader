package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type sshKeyDeleter interface {
	Delete() error
}

type Rotate struct {
	stateValidator stateValidator
	sshKeyDeleter  sshKeyDeleter
	up             up
}

func NewRotate(stateValidator stateValidator, sshKeyDeleter sshKeyDeleter, up up) Rotate {
	return Rotate{
		stateValidator: stateValidator,
		sshKeyDeleter:  sshKeyDeleter,
		up:             up,
	}
}

func (r Rotate) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := r.stateValidator.Validate()
	if err != nil {
		return fmt.Errorf("validate state: %s", err)
	}

	err = r.up.CheckFastFails(subcommandFlags, state)
	if err != nil {
		return fmt.Errorf("up: %s", err)
	}
	return nil
}

func (r Rotate) Execute(args []string, state storage.State) error {
	err := r.sshKeyDeleter.Delete()
	if err != nil {
		return fmt.Errorf("delete ssh key: %s", err)
	}

	err = r.up.Execute(args, state)
	if err != nil {
		return fmt.Errorf("up: %s", err)
	}

	return nil
}
