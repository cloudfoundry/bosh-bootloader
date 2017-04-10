package commands

import "github.com/cloudfoundry/bosh-bootloader/helpers"

func handleTerraformError(err error, stateStore stateStore) error {
	switch err.(type) {
	case terraformManagerError:
		terraformManagerError := err.(terraformManagerError)
		updatedBBLState, bblStateErr := terraformManagerError.BBLState()
		if bblStateErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(bblStateErr)
			return errorList
		}
		setErr := stateStore.Set(updatedBBLState)
		if setErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(setErr)
			return errorList
		}
	}

	return err
}
