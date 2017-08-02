package application

import "fmt"

type CredentialValidator struct {
	iaas                   string
	awsCredentialValidator credentialValidator
	gcpCredentialValidator credentialValidator
}

type credentialValidator interface {
	Validate() error
}

func NewCredentialValidator(iaas string, gcpCredentialValidator credentialValidator, awsCredentialValidator credentialValidator) CredentialValidator {
	return CredentialValidator{
		iaas: iaas,
		awsCredentialValidator: awsCredentialValidator,
		gcpCredentialValidator: gcpCredentialValidator,
	}
}

func (c CredentialValidator) Validate() error {
	switch c.iaas {
	case "aws":
		return c.awsCredentialValidator.Validate()
	case "gcp":
		return c.gcpCredentialValidator.Validate()
	default:
		return fmt.Errorf("cannot validate credentials: invalid iaas %q", c.iaas)
	}
	return nil
}
