package iam

import (
	"fmt"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/aws/common"
)

type rolesClient interface {
	ListRoles(*awsiam.ListRolesInput) (*awsiam.ListRolesOutput, error)
	DeleteRole(*awsiam.DeleteRoleInput) (*awsiam.DeleteRoleOutput, error)
}

type Roles struct {
	client   rolesClient
	logger   logger
	policies rolePolicies
}

func NewRoles(client rolesClient, logger logger, policies rolePolicies) Roles {
	return Roles{
		client:   client,
		logger:   logger,
		policies: policies,
	}
}

func (r Roles) ListOnly(filter string) ([]common.Deletable, error) {
	return r.getRoles(filter)
}

func (o Roles) List(filter string) ([]common.Deletable, error) {
	resources, err := o.getRoles(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := o.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (r Roles) getRoles(filter string) ([]common.Deletable, error) {
	roles, err := r.client.ListRoles(&awsiam.ListRolesInput{})
	if err != nil {
		return nil, fmt.Errorf("List IAM Roles: %s", err)
	}

	var resources []common.Deletable
	for _, role := range roles.Roles {
		resource := NewRole(r.client, r.policies, role.RoleName)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
