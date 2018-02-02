package iam

import (
	"fmt"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type User struct {
	client     usersClient
	policies   userPolicies
	accessKeys accessKeys
	name       *string
	identifier string
}

func NewUser(client usersClient, policies userPolicies, accessKeys accessKeys, name *string) User {
	return User{
		client:     client,
		policies:   policies,
		accessKeys: accessKeys,
		name:       name,
		identifier: *name,
	}
}

func (u User) Delete() error {
	err := u.accessKeys.Delete(*u.name)
	if err != nil {
		return fmt.Errorf("FAILED deleting access keys for %s: %s", u.identifier, err)
	}

	err = u.policies.Delete(*u.name)
	if err != nil {
		return fmt.Errorf("FAILED deleting policies for %s: %s", u.identifier, err)
	}

	_, err = u.client.DeleteUser(&awsiam.DeleteUserInput{
		UserName: u.name,
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting user %s: %s", u.identifier, err)
	}

	return err
}
