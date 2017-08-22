package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type sshKeyDeleter interface {
	Delete(storage.State) (storage.State, error)
}

type up interface {
	CheckFastFails([]string, storage.State) error
	Execute([]string, storage.State) error
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
	updatedState, err := r.sshKeyDeleter.Delete(state)
	if err != nil {
		return fmt.Errorf("delete ssh key: %s", err)
	}

	err = r.up.Execute(args, updatedState)
	if err != nil {
		return fmt.Errorf("up: %s", err)
	}

	return nil
}
