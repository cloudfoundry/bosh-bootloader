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
	rtype      string
}

func NewUser(client usersClient, policies userPolicies, accessKeys accessKeys, name *string) User {
	return User{
		client:     client,
		policies:   policies,
		accessKeys: accessKeys,
		name:       name,
		identifier: *name,
		rtype:      "IAM User",
	}
}

func (u User) Delete() error {
	err := u.accessKeys.Delete(u.identifier)
	if err != nil {
		return fmt.Errorf("Delete access keys: %s", err)
	}

	err = u.policies.Delete(u.identifier)
	if err != nil {
		return fmt.Errorf("Delete policies: %s", err)
	}

	_, err = u.client.DeleteUser(&awsiam.DeleteUserInput{UserName: u.name})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return err
}

func (u User) Name() string {
	return u.identifier
}

func (u User) Type() string {
	return u.rtype
}
