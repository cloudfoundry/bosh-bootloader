package aws

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/application"
)

type CredentialValidator struct {
	configuration application.Configuration
}

func NewCredentialValidator(configuration application.Configuration) CredentialValidator {
	return CredentialValidator{
		configuration: configuration,
	}
}

func (c CredentialValidator) Validate() error {
	if c.configuration.State.AWS.AccessKeyID == "" {
		return errors.New("AWS access key ID must be provided")
	}

	if c.configuration.State.AWS.SecretAccessKey == "" {
		return errors.New("AWS secret access key must be provided")
	}

	if c.configuration.State.AWS.Region == "" {
		return errors.New("AWS region must be provided")
	}

	return nil
}
