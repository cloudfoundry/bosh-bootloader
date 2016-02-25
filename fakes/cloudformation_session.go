package fakes

import "github.com/aws/aws-sdk-go/service/cloudformation"

type CloudFormationSession struct{}

func (c *CloudFormationSession) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	return nil, nil
}

func (c *CloudFormationSession) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	return nil, nil
}

func (c *CloudFormationSession) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	return nil, nil
}
