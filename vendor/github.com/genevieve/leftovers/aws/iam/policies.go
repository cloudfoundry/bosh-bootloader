package iam

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/common"
)

type policiesClient interface {
	ListPolicies(*awsiam.ListPoliciesInput) (*awsiam.ListPoliciesOutput, error)
	ListPolicyVersions(*awsiam.ListPolicyVersionsInput) (*awsiam.ListPolicyVersionsOutput, error)
	DeletePolicyVersion(*awsiam.DeletePolicyVersionInput) (*awsiam.DeletePolicyVersionOutput, error)
	DeletePolicy(*awsiam.DeletePolicyInput) (*awsiam.DeletePolicyOutput, error)
}

type Policies struct {
	client policiesClient
	logger logger
}

func NewPolicies(client policiesClient, logger logger) Policies {
	return Policies{
		client: client,
		logger: logger,
	}
}

func (p Policies) List(filter string) ([]common.Deletable, error) {
	policies, err := p.client.ListPolicies(&awsiam.ListPoliciesInput{Scope: aws.String("Local")})
	if err != nil {
		return nil, fmt.Errorf("List IAM Policies: %s", err)
	}

	var resources []common.Deletable
	for _, o := range policies.Policies {
		r := NewPolicy(p.client, p.logger, o.PolicyName, o.Arn)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := p.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (p Policies) Type() string {
	return "iam-policy"
}
