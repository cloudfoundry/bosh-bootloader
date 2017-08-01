package gcp

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type EnvironmentValidator struct{}

func NewEnvironmentValidator() EnvironmentValidator {
	return EnvironmentValidator{}
}

func (e EnvironmentValidator) Validate(state storage.State) error {
	if state.TFState == "" {
		return application.BBLNotFound
	}

	return nil
}
