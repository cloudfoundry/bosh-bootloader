package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cloudfoundry/bosh-bootloader/aws"

	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
)

type Client interface {
	CreateStack(input *awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error)
	UpdateStack(input *awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error)
	DescribeStacks(input *awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error)
	DeleteStack(input *awscloudformation.DeleteStackInput) (*awscloudformation.DeleteStackOutput, error)
	DescribeStackResource(input *awscloudformation.DescribeStackResourceInput) (*awscloudformation.DescribeStackResourceOutput, error)
}

func NewClient(config aws.Config) Client {
	return awscloudformation.New(session.New(config.ClientConfig()))
}
