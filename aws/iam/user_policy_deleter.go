package iam

import (
	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type UserPolicyDeleter struct {
	client Client
}

func NewUserPolicyDeleter(client Client) UserPolicyDeleter {
	return UserPolicyDeleter{
		client: client,
	}
}

func (c UserPolicyDeleter) Delete(username, policyName string) error {
	_, err := c.client.DeleteUserPolicy(&awsiam.DeleteUserPolicyInput{
		UserName:   aws.String(username),
		PolicyName: aws.String(policyName),
	})

	return err
}
