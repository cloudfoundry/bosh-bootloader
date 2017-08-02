package gcp

import (
	"errors"
)

type CredentialValidator struct {
	projectID         string
	serviceAccountKey string
	region            string
	zone              string
}

func NewCredentialValidator(projectID, serviceAccountKey, region, zone string) CredentialValidator {
	return CredentialValidator{
		projectID:         projectID,
		serviceAccountKey: serviceAccountKey,
		region:            region,
		zone:              zone,
	}
}

func (c CredentialValidator) Validate() error {
	if c.projectID == "" {
		return errors.New("GCP project ID must be provided")
	}

	if c.serviceAccountKey == "" {
		return errors.New("GCP service account key must be provided")
	}

	if c.region == "" {
		return errors.New("GCP region must be provided")
	}

	if c.zone == "" {
		return errors.New("GCP zone must be provided")
	}

	return nil
}
