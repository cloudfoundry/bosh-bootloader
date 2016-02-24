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
}

func (c *CloudFormationClient) CreateStack(input *awscloudformation.CreateStackInput) (*awscloudformation.CreateStackOutput, error) {
	c.CreateStackCall.Receives.CreateStackInput = input
	return nil, c.CreateStackCall.Returns.Error
}
