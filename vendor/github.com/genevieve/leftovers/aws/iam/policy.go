package iam

import (
	"fmt"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type Policy struct {
	client     policiesClient
	name       *string
	arn        *string
	identifier string
}

func NewPolicy(client policiesClient, name, arn *string) Policy {
	return Policy{
		client:     client,
		name:       name,
		arn:        arn,
		identifier: *name,
	}
}

func (p Policy) Delete() error {
	versions, err := p.client.ListPolicyVersions(&awsiam.ListPolicyVersionsInput{PolicyArn: p.arn})
	if err != nil {
		return fmt.Errorf("FAILED listing versions for policy %s: %s", p.identifier, err)
	}

	for _, v := range versions.Versions {
		if !*v.IsDefaultVersion {
			_, err := p.client.DeletePolicyVersion(&awsiam.DeletePolicyVersionInput{
				PolicyArn: p.arn,
				VersionId: v.VersionId,
			})
			if err != nil {
				return fmt.Errorf("FAILED deleting version %s of policy %s: %s", *v.VersionId, p.identifier, err)
			}
		}
	}

	_, err = p.client.DeletePolicy(&awsiam.DeletePolicyInput{
		PolicyArn: p.arn,
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting policy %s: %s", p.identifier, err)
	}

	return nil
}

func (p Policy) Name() string {
	return p.identifier
}
