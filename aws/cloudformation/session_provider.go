package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws/session"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
)

type SessionProvider struct{}

type Session interface {
	CreateStack(input *awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error)
	UpdateStack(input *awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error)
	DescribeStacks(input *awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error)
}

func NewSessionProvider() SessionProvider {
	return SessionProvider{}
}

func (s SessionProvider) Session(config aws.Config) (Session, error) {
	if err := config.ValidateCredentials(); err != nil {
		return nil, err
	}
	return awscloudformation.New(session.New(config.SessionConfig())), nil
}
