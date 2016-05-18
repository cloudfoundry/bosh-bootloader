package application

import "errors"

type AWSCredentialValidator struct{}

func NewAWSCredentialValidator() AWSCredentialValidator {
	return AWSCredentialValidator{}
}

func (AWSCredentialValidator) Validate(accessKeyID string, secretAccessKey string, region string) error {
	if accessKeyID == "" {
		return errors.New("--aws-access-key-id must be provided")
	}

	if secretAccessKey == "" {
		return errors.New("--aws-secret-access-key must be provided")
	}

	if region == "" {
		return errors.New("--aws-region must be provided")
	}

	return nil
}
