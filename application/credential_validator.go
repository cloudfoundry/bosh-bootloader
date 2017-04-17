package application

import "fmt"

type CredentialValidator struct {
	configuration          Configuration
	awsCredentialValidator credentialValidator
	gcpCredentialValidator credentialValidator
}

type credentialValidator interface {
	Validate() error
}

func NewCredentialValidator(configuration Configuration, gcpCredentialValidator credentialValidator, awsCredentialValidator credentialValidator) CredentialValidator {
	return CredentialValidator{
		configuration:          configuration,
		awsCredentialValidator: awsCredentialValidator,
		gcpCredentialValidator: gcpCredentialValidator,
	}
}

func (c CredentialValidator) Validate() error {
	switch c.configuration.State.IAAS {
	case "aws":
		return c.awsCredentialValidator.Validate()
	case "gcp":
		return c.gcpCredentialValidator.Validate()
	default:
		return fmt.Errorf("cannot validate credentials: invalid iaas %q", c.configuration.State.IAAS)
	}
	return nil
}
