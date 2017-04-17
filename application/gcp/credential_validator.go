package gcp

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
