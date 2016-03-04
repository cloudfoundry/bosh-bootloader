package aws

import (
	"errors"

	goaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type Config struct {
	AccessKeyID      string
	SecretAccessKey  string
	Region           string
	EndpointOverride string
}

func (c Config) ValidateCredentials() error {
	if c.AccessKeyID == "" {
		return errors.New("--aws-access-key-id must be provided")
	}

	if c.SecretAccessKey == "" {
		return errors.New("--aws-secret-access-key must be provided")
	}

	if c.Region == "" {
		return errors.New("--aws-region must be provided")
	}

	return nil
}

func (c Config) ClientConfig() *goaws.Config {
	awsConfig := &goaws.Config{
		Credentials: credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""),
		Region:      goaws.String(c.Region),
	}

	if c.EndpointOverride != "" {
		awsConfig.WithEndpoint(c.EndpointOverride)
	}

	return awsConfig
}
