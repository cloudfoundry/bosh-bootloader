package iam

import (
	"fmt"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/aws/common"
)

type usersClient interface {
	ListUsers(*awsiam.ListUsersInput) (*awsiam.ListUsersOutput, error)
	DeleteUser(*awsiam.DeleteUserInput) (*awsiam.DeleteUserOutput, error)
}

type Users struct {
	client     usersClient
	logger     logger
	policies   userPolicies
	accessKeys accessKeys
}

func NewUsers(client usersClient, logger logger, policies userPolicies, accessKeys accessKeys) Users {
	return Users{
		client:     client,
		logger:     logger,
		policies:   policies,
		accessKeys: accessKeys,
	}
}

func (u Users) ListOnly(filter string) ([]common.Deletable, error) {
	return u.getUsers(filter)
}

func (u Users) List(filter string) ([]common.Deletable, error) {
	resources, err := u.getUsers(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := u.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (u Users) getUsers(filter string) ([]common.Deletable, error) {
	users, err := u.client.ListUsers(&awsiam.ListUsersInput{})
	if err != nil {
		return nil, fmt.Errorf("List IAM Users: %s", err)
	}

	var resources []common.Deletable
	for _, r := range users.Users {
		resource := NewUser(u.client, u.policies, u.accessKeys, r.UserName)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
