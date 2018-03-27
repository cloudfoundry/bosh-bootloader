package iam

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/aws/common"
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

func (p Policies) ListAll(filter string) ([]common.Deletable, error) {
	return p.getPolicies(filter)
}

func (p Policies) List(filter string) ([]common.Deletable, error) {
	resources, err := p.getPolicies(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := p.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (p Policies) getPolicies(filter string) ([]common.Deletable, error) {
	policies, err := p.client.ListPolicies(&awsiam.ListPoliciesInput{Scope: aws.String("Local")})
	if err != nil {
		return nil, fmt.Errorf("Listing policies: %s", err)
	}

	var resources []common.Deletable
	for _, o := range policies.Policies {
		resource := NewPolicy(p.client, o.PolicyName, o.Arn)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
