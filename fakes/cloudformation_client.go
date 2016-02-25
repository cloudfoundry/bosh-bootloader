package fakes

import (
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
)

type CloudFormationClient struct {
	CreateStackCall struct {
		Receives struct {
			CreateStackInput *awscloudformation.CreateStackInput
		}
		Returns struct {
			Error error
		}
	}

	UpdateStackCall struct {
		Receives struct {
			UpdateStackInput *awscloudformation.UpdateStackInput
		}
		Returns struct {
			Error error
		}
	}

	DescribeStacksCall struct {
		Receives struct {
			Input *awscloudformation.DescribeStacksInput
		}
		Returns struct {
			Output *awscloudformation.DescribeStacksOutput
			Error  error
		}
	}
}

func (c *CloudFormationClient) CreateStack(input *awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error) {
	c.CreateStackCall.Receives.CreateStackInput = input
	return nil, c.CreateStackCall.Returns.Error
}

func (c *CloudFormationClient) UpdateStack(input *awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error) {
	c.UpdateStackCall.Receives.UpdateStackInput = input
	return nil, c.UpdateStackCall.Returns.Error
}

func (c *CloudFormationClient) DescribeStacks(input *awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error) {
	c.DescribeStacksCall.Receives.Input = input
	return c.DescribeStacksCall.Returns.Output, c.DescribeStacksCall.Returns.Error
}
