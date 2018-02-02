package iam

import (
	"fmt"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type Role struct {
	client     rolesClient
	policies   rolePolicies
	name       *string
	identifier string
}

func NewRole(client rolesClient, policies rolePolicies, name *string) Role {
	return Role{
		client:     client,
		policies:   policies,
		name:       name,
		identifier: *name,
	}
}

func (r Role) Delete() error {
	err := r.policies.Delete(*r.name)
	if err != nil {
		return fmt.Errorf("FAILED deleting policies for %s: %s", r.identifier, err)
	}

	_, err = r.client.DeleteRole(&awsiam.DeleteRoleInput{
		RoleName: r.name,
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting role %s: %s", r.identifier, err)
	}

	return nil
}
