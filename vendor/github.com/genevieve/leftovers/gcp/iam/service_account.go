package iam

import "fmt"

type ServiceAccount struct {
	client serviceAccountsClient
	name   string
}

func NewServiceAccount(client serviceAccountsClient, name string) ServiceAccount {
	return ServiceAccount{
		client: client,
		name:   name,
	}
}

func (s ServiceAccount) Delete() error {
	err := s.client.DeleteServiceAccount(s.name)
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (s ServiceAccount) Name() string {
	return s.name
}

func (s ServiceAccount) Type() string {
	return "IAM Service Account"
}
