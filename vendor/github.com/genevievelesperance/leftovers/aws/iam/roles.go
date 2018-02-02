package iam

import (
	"fmt"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevievelesperance/leftovers/aws/common"
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

func (r Roles) List(filter string) ([]common.Deletable, error) {
	roles, err := r.client.ListRoles(&awsiam.ListRolesInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing roles: %s", err)
	}

	var resources []common.Deletable
	for _, role := range roles.Roles {
		resource := NewRole(r.client, r.policies, role.RoleName)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := r.logger.Prompt(fmt.Sprintf("Are you sure you want to delete role %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
