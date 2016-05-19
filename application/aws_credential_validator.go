package application

import "errors"

type AWSCredentialValidator struct {
	configuration Configuration
}

func NewAWSCredentialValidator(configuration Configuration) AWSCredentialValidator {
	return AWSCredentialValidator{
		configuration: configuration,
	}
}

func (a AWSCredentialValidator) Validate() error {
	if a.configuration.State.AWS.AccessKeyID == "" {
		return errors.New("--aws-access-key-id must be provided")
	}

	if a.configuration.State.AWS.SecretAccessKey == "" {
		return errors.New("--aws-secret-access-key must be provided")
	}

	if a.configuration.State.AWS.Region == "" {
		return errors.New("--aws-region must be provided")
	}

	return nil
}
