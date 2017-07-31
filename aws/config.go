package aws

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

func (c Config) ClientConfig() *goaws.Config {
	awsConfig := &goaws.Config{
		Credentials: credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""),
		Region:      goaws.String(c.Region),
	}

	return awsConfig
}
