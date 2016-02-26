package fakes

import (
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
)

type CloudFormationClient struct {
	CreateStackCall struct {
		Receives struct {
			Input *awscloudformation.CreateStackInput
		}
		Returns struct {
			Error error
		}
	}

	UpdateStackCall struct {
		Receives struct {
			Input *awscloudformation.UpdateStackInput
		}
		Returns struct {
			Error error
		}
	}

	DescribeStacksCall struct {
		CallCount int
		Stub      func(*awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error)

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
	c.CreateStackCall.Receives.Input = input
	return nil, c.CreateStackCall.Returns.Error
}

func (c *CloudFormationClient) UpdateStack(input *awscloudformation.UpdateStackInput) (*awscloudformation.UpdateStackOutput, error) {
	c.UpdateStackCall.Receives.Input = input
	return nil, c.UpdateStackCall.Returns.Error
}

func (c *CloudFormationClient) DescribeStacks(input *awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error) {
	c.DescribeStacksCall.CallCount++
	c.DescribeStacksCall.Receives.Input = input

	if c.DescribeStacksCall.Stub != nil {
		return c.DescribeStacksCall.Stub(input)
	}

	return c.DescribeStacksCall.Returns.Output, c.DescribeStacksCall.Returns.Error
}
