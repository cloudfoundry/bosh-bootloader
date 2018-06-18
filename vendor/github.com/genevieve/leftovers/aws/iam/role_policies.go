package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type rolePoliciesClient interface {
	ListAttachedRolePolicies(*awsiam.ListAttachedRolePoliciesInput) (*awsiam.ListAttachedRolePoliciesOutput, error)
	ListRolePolicies(*awsiam.ListRolePoliciesInput) (*awsiam.ListRolePoliciesOutput, error)
	DetachRolePolicy(*awsiam.DetachRolePolicyInput) (*awsiam.DetachRolePolicyOutput, error)
	DeleteRolePolicy(*awsiam.DeleteRolePolicyInput) (*awsiam.DeleteRolePolicyOutput, error)
}

type rolePolicies interface {
	Delete(roleName string) error
}

type RolePolicies struct {
	client rolePoliciesClient
	logger logger
}

func NewRolePolicies(client rolePoliciesClient, logger logger) RolePolicies {
	return RolePolicies{
		client: client,
		logger: logger,
	}
}

func (o RolePolicies) Delete(roleName string) error {
	attachedPolicies, err := o.client.ListAttachedRolePolicies(&awsiam.ListAttachedRolePoliciesInput{RoleName: aws.String(roleName)})
	if err != nil {
		return fmt.Errorf("List IAM Attached Role Policies: %s", err)
	}

	for _, p := range attachedPolicies.AttachedPolicies {
		n := *p.PolicyName

		_, err := o.client.DetachRolePolicy(&awsiam.DetachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: p.PolicyArn,
		})
		if err == nil {
			o.logger.Printf("[IAM Role: %s] Detached policy %s \n", roleName, n)
		} else {
			o.logger.Printf("[IAM Role: %s] Detach policy %s: %s \n", roleName, n, err)
		}

		_, err = o.client.DeleteRolePolicy(&awsiam.DeleteRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: p.PolicyName,
		})
		if err == nil {
			o.logger.Printf("[IAM Role: %s] Deleted policy %s \n", roleName, n)
		} else {
			o.logger.Printf("[IAM Role: %s] Delete policy %s: %s \n", roleName, n, err)
		}
	}

	policies, err := o.client.ListRolePolicies(&awsiam.ListRolePoliciesInput{RoleName: aws.String(roleName)})
	if err != nil {
		return fmt.Errorf("List IAM Role Policies: %s", err)
	}

	for _, p := range policies.PolicyNames {
		n := *p

		_, err = o.client.DeleteRolePolicy(&awsiam.DeleteRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: p,
		})
		if err == nil {
			o.logger.Printf("[IAM Role: %s] Deleted policy %s \n", roleName, n)
		} else {
			o.logger.Printf("[IAM Role: %s] Delete policy %s: %s \n", roleName, n, err)
		}
	}

	return nil
}
