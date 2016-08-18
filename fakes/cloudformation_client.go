package fakes

import "github.com/aws/aws-sdk-go/service/cloudformation"

type CloudFormationClient struct {
	CreateStackCall struct {
		Receives struct {
			Input *cloudformation.CreateStackInput
		}
		Returns struct {
			Error error
		}
	}

	UpdateStackCall struct {
		CallCount int
		Receives  struct {
			Input *cloudformation.UpdateStackInput
		}
		Returns struct {
			Error error
		}
	}

	DescribeStacksCall struct {
		CallCount int
		Stub      func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)

		Receives struct {
			Input *cloudformation.DescribeStacksInput
		}
		Returns struct {
			Output *cloudformation.DescribeStacksOutput
			Error  error
		}
	}

	DeleteStackCall struct {
		Receives struct {
			Input *cloudformation.DeleteStackInput
		}
		Returns struct {
			Output *cloudformation.DeleteStackOutput
			Error  error
		}
	}

	DescribeStackResourcesCall struct {
		Receives struct {
			Input *cloudformation.DescribeStackResourcesInput
		}
		Returns struct {
			Output *cloudformation.DescribeStackResourcesOutput
			Error  error
		}
	}
}

func (c *CloudFormationClient) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	c.CreateStackCall.Receives.Input = input
	return nil, c.CreateStackCall.Returns.Error
}

func (c *CloudFormationClient) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	c.UpdateStackCall.CallCount++
	c.UpdateStackCall.Receives.Input = input
	return nil, c.UpdateStackCall.Returns.Error
}

func (c *CloudFormationClient) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	c.DescribeStacksCall.CallCount++
	c.DescribeStacksCall.Receives.Input = input

	if c.DescribeStacksCall.Stub != nil {
		return c.DescribeStacksCall.Stub(input)
	}

	return c.DescribeStacksCall.Returns.Output, c.DescribeStacksCall.Returns.Error
}

func (c *CloudFormationClient) DeleteStack(input *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	c.DeleteStackCall.Receives.Input = input
	return c.DeleteStackCall.Returns.Output, c.DeleteStackCall.Returns.Error
}

func (c *CloudFormationClient) DescribeStackResources(input *cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error) {
	c.DescribeStackResourcesCall.Receives.Input = input
	return c.DescribeStackResourcesCall.Returns.Output, c.DescribeStackResourcesCall.Returns.Error

}
