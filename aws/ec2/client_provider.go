package ec2

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
)

type ClientProvider struct{}

type Client interface {
	ImportKeyPair(*ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error)
	DescribeKeyPairs(*ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error)
	CreateKeyPair(*ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error)
}

func NewClientProvider() ClientProvider {
	return ClientProvider{}
}

func (s ClientProvider) Client(config aws.Config) (Client, error) {
	if err := config.ValidateCredentials(); err != nil {
		return nil, err
	}

	return ec2.New(session.New(config.SessionConfig())), nil
}
