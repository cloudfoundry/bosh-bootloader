package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	awselb "github.com/aws/aws-sdk-go/service/elb"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/elb"
)

type ClientProvider struct{}

func NewClientProvider() ClientProvider {
	return ClientProvider{}
}

func (s ClientProvider) ELBClient(config Config) (elb.Client, error) {
	if err := config.ValidateCredentials(); err != nil {
		return nil, err
	}
	return awselb.New(session.New(config.ClientConfig())), nil
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
