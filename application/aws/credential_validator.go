package aws

import (
	"errors"
)

type CredentialValidator struct {
	accessKeyID     string
	secretAccessKey string
	region          string
}

func NewCredentialValidator(accessKeyID, secretAccessKey, region string) CredentialValidator {
	return CredentialValidator{
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region:          region,
	}
}

func (c CredentialValidator) Validate() error {
	if c.accessKeyID == "" {
		return errors.New("AWS access key ID must be provided")
	}

	if c.secretAccessKey == "" {
		return errors.New("AWS secret access key must be provided")
	}

	if c.region == "" {
		return errors.New("AWS region must be provided")
	}

	return nil
}
