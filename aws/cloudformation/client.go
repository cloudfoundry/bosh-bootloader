package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"

	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
)

type Client interface {
	CreateStack(input *awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error)
	UpdateStack(input *awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error)
	DescribeStacks(input *awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error)
	DeleteStack(input *awscloudformation.DeleteStackInput) (*awscloudformation.DeleteStackOutput, error)
}

func NewClient(config aws.Config) Client {
	return awscloudformation.New(session.New(config.ClientConfig()))
}
