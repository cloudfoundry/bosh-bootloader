package fakes

import "github.com/aws/aws-sdk-go/service/cloudformation"

type CloudFormationSession struct{}

func (c *CloudFormationSession) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	return nil, nil
}
