package aws

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type BrokenEnvironmentValidator struct {
	infrastructureManager infrastructureManager
}

func NewBrokenEnvironmentValidator(infrastructureManager infrastructureManager) BrokenEnvironmentValidator {
	return BrokenEnvironmentValidator{
		infrastructureManager: infrastructureManager,
	}
}

func (b BrokenEnvironmentValidator) Validate(state storage.State) error {
	if state.TFState != "" {
		return nil
	}

	stackExists, err := b.infrastructureManager.Exists(state.Stack.Name)
	if err != nil {
		return err
	}

	if !stackExists && !state.BOSH.IsEmpty() {
		return fmt.Errorf("Found BOSH data in state directory, "+
			"but Cloud Formation stack %q cannot be found for region %q and given "+
			"AWS credentials. bbl cannot safely proceed. Open an issue on GitHub at "+
			"https://github.com/cloudfoundry/bosh-bootloader/issues/new if you need assistance.",
			state.Stack.Name, state.AWS.Region)
	}

	return nil
}
