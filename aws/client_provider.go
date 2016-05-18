package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	awselb "github.com/aws/aws-sdk-go/service/elb"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/elb"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
)

type ClientProvider struct{}

func NewClientProvider() ClientProvider {
	return ClientProvider{}
}

func (s ClientProvider) ELBClient(config Config) elb.Client {
	return awselb.New(session.New(config.ClientConfig()))
}

func (s ClientProvider) CloudFormationClient(config Config) cloudformation.Client {
	return awscloudformation.New(session.New(config.ClientConfig()))
}

func (s ClientProvider) EC2Client(config Config) ec2.Client {
	return awsec2.New(session.New(config.ClientConfig()))
}

func (s ClientProvider) IAMClient(config Config) iam.Client {
	return awsiam.New(session.New(config.ClientConfig()))
}
