package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

func handleTerraformError(err error, state storage.State, stateStore stateStore) error {
	errorList := helpers.Errors{}
	errorList.Add(err)

	setErr := stateStore.Set(state)
	if setErr != nil {
		errorList.Add(setErr)
	}

	return errors.New(errorList.Error())
}

type ExitSuccessfully struct{}

func (e ExitSuccessfully) Error() string {
	return "Succeeded, exiting early"
}

type NoBBLStateError struct {
	dir string
}

func NewNoBBLStateError(dir string) NoBBLStateError {
	return NoBBLStateError{dir: dir}
}

func (e NoBBLStateError) Error() string {
	return fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", e.dir)
}

func (e NoBBLStateError) String() string {
	return e.Error()
}
