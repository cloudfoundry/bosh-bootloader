package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
)

type ClientProvider struct{}

func NewClientProvider() ClientProvider {
	return ClientProvider{}
}

func (s ClientProvider) CloudFormationClient(config Config) (cloudformation.Client, error) {
	if err := config.ValidateCredentials(); err != nil {
		return nil, err
	}
	return awscloudformation.New(session.New(config.ClientConfig())), nil
}

func (s ClientProvider) EC2Client(config Config) (ec2.Client, error) {
	if err := config.ValidateCredentials(); err != nil {
		return nil, err
	}

	return awsec2.New(session.New(config.ClientConfig())), nil
}
