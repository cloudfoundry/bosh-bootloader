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
		return errors.New("AWS access key ID must be provided")
	}

	if a.configuration.State.AWS.SecretAccessKey == "" {
		return errors.New("AWS secret access key must be provided")
	}

	if a.configuration.State.AWS.Region == "" {
		return errors.New("AWS region must be provided")
	}

	return nil
}
