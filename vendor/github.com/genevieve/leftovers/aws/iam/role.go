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
	rtype      string
}

func NewRole(client rolesClient, policies rolePolicies, name *string) Role {
	return Role{
		client:     client,
		policies:   policies,
		name:       name,
		identifier: *name,
		rtype:      "IAM Role",
	}
}

func (r Role) Delete() error {
	err := r.policies.Delete(r.identifier)
	if err != nil {
		return fmt.Errorf("Delete policies: %s", err)
	}

	_, err = r.client.DeleteRole(&awsiam.DeleteRoleInput{RoleName: r.name})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (r Role) Name() string {
	return r.identifier
}

func (r Role) Type() string {
	return r.rtype
}
