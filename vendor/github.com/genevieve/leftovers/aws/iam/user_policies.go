package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
		return fmt.Errorf("List IAM User Policies: %s", err)
	}

	for _, p := range policies.AttachedPolicies {
		n := *p.PolicyName

		_, err = o.client.DetachUserPolicy(&awsiam.DetachUserPolicyInput{
			UserName:  aws.String(userName),
			PolicyArn: p.PolicyArn,
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchEntity" {
				o.logger.Printf("[IAM User: %s] Detached policy %s \n", userName, n)
			} else {
				o.logger.Printf("[IAM User: %s] Detach policy %s: %s \n", userName, n, err)
			}
		} else {
			o.logger.Printf("[IAM User: %s] Detached policy %s \n", userName, n)
		}

		_, err = o.client.DeleteUserPolicy(&awsiam.DeleteUserPolicyInput{
			UserName:   aws.String(userName),
			PolicyName: p.PolicyName,
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchEntity" {
				o.logger.Printf("[IAM User: %s] Deleted policy %s \n", userName, n)
			} else {
				o.logger.Printf("[IAM User: %s] Delete policy %s: %s \n", userName, n, err)
			}
		} else {
			o.logger.Printf("[IAM User: %s] Deleted policy %s \n", userName, n)
		}
	}

	return nil
}
