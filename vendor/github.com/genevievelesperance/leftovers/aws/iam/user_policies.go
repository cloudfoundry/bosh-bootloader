package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type userPoliciesClient interface {
	ListAttachedUserPolicies(*awsiam.ListAttachedUserPoliciesInput) (*awsiam.ListAttachedUserPoliciesOutput, error)
	DetachUserPolicy(*awsiam.DetachUserPolicyInput) (*awsiam.DetachUserPolicyOutput, error)
	DeleteUserPolicy(*awsiam.DeleteUserPolicyInput) (*awsiam.DeleteUserPolicyOutput, error)
}

type userPolicies interface {
	Delete(userName string) error
}

type UserPolicies struct {
	client userPoliciesClient
	logger logger
}

func NewUserPolicies(client userPoliciesClient, logger logger) UserPolicies {
	return UserPolicies{
		client: client,
		logger: logger,
	}
}

func (o UserPolicies) Delete(userName string) error {
	policies, err := o.client.ListAttachedUserPolicies(&awsiam.ListAttachedUserPoliciesInput{UserName: aws.String(userName)})
	if err != nil {
		return fmt.Errorf("Listing user policies: %s", err)
	}

	for _, p := range policies.AttachedPolicies {
		n := *p.PolicyName

		_, err = o.client.DetachUserPolicy(&awsiam.DetachUserPolicyInput{
			UserName:  aws.String(userName),
			PolicyArn: p.PolicyArn,
		})
		if err == nil {
			o.logger.Printf("SUCCESS detaching user policy %s\n", n)
		} else {
			o.logger.Printf("ERROR detaching user policy %s: %s\n", n, err)
		}

		_, err = o.client.DeleteUserPolicy(&awsiam.DeleteUserPolicyInput{
			UserName:   aws.String(userName),
			PolicyName: p.PolicyName,
		})
		if err == nil {
			o.logger.Printf("SUCCESS deleting user policy %s\n", n)
		} else {
			o.logger.Printf("ERROR deleting user policy %s: %s\n", n, err)
		}
	}

	return nil
}
