package application

import "errors"

type CredentialValidator struct {
	configuration Configuration
}

func NewCredentialValidator(configuration Configuration) CredentialValidator {
	return CredentialValidator{
		configuration: configuration,
	}
}

func (c CredentialValidator) ValidateAWS() error {
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

func (c CredentialValidator) ValidateGCP() error {
	if c.configuration.State.GCP.ProjectID == "" {
		return errors.New("GCP project ID must be provided")
	}

	if c.configuration.State.GCP.ServiceAccountKey == "" {
		return errors.New("GCP service account key must be provided")
	}

	if c.configuration.State.GCP.Region == "" {
		return errors.New("GCP region must be provided")
	}

	if c.configuration.State.GCP.Zone == "" {
		return errors.New("GCP zone must be provided")
	}

	return nil
}
