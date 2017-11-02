package commands

import (
	"errors"

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
