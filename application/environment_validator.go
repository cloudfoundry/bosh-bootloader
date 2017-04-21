package application

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var BBLNotFound error = errors.New("a bbl environment could not be found, please create a new environment before running this command again")

type EnvironmentValidator struct {
	awsEnvironmentValidator environmentValidator
	gcpEnvironmentValidator environmentValidator
}

type environmentValidator interface {
	Validate(storage.State) error
}

func NewEnvironmentValidator(awsEnvironmentValidator environmentValidator, gcpEnvironmentValidator environmentValidator) EnvironmentValidator {
	return EnvironmentValidator{
		awsEnvironmentValidator: awsEnvironmentValidator,
		gcpEnvironmentValidator: gcpEnvironmentValidator,
	}
}

func (e EnvironmentValidator) Validate(state storage.State) error {
	switch state.IAAS {
	case "gcp":
		return e.gcpEnvironmentValidator.Validate(state)
	case "aws":
		return e.awsEnvironmentValidator.Validate(state)
	default:
		return fmt.Errorf("invalid IAAS specified: %s", state.IAAS)
	}
}
