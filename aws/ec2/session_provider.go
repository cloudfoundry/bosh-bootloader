package ec2

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type SessionProvider struct{}

type Session interface {
	ImportKeyPair(input *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error)
}

type Config struct {
	AccessKeyID      string
	SecretAccessKey  string
	Region           string
	EndpointOverride string
}

func NewSessionProvider() SessionProvider {
	return SessionProvider{}
}

func (s SessionProvider) Session(config Config) Session {
	awsConfig := &goaws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		Region:      goaws.String(config.Region),
	}

	if config.EndpointOverride != "" {
		awsConfig.WithEndpoint(config.EndpointOverride)
	}

	return ec2.New(session.New(awsConfig))
}
