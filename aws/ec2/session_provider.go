package ec2

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
)

type SessionProvider struct{}

type Session interface {
	ImportKeyPair(*ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error)
	DescribeKeyPairs(*ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error)
	CreateKeyPair(*ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error)
}

func NewSessionProvider() SessionProvider {
	return SessionProvider{}
}

func (s SessionProvider) Session(config aws.Config) (Session, error) {
	if err := config.ValidateCredentials(); err != nil {
		return nil, err
	}

	return ec2.New(session.New(config.SessionConfig())), nil
}
