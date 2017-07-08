package iam

import (
	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type UserPolicyDeleter struct {
	iamClientProvider iamClientProvider
}

func NewUserPolicyDeleter(iamClientProvider iamClientProvider) UserPolicyDeleter {
	return UserPolicyDeleter{
		iamClientProvider: iamClientProvider,
	}
}

func (c UserPolicyDeleter) Delete(username, policyName string) error {
	_, err := c.iamClientProvider.GetIAMClient().DeleteUserPolicy(&awsiam.DeleteUserPolicyInput{
		UserName:   aws.String(username),
		PolicyName: aws.String(policyName),
	})

	return err
}
