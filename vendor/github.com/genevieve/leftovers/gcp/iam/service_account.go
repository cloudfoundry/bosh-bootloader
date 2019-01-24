package iam

import (
	"fmt"
	"strings"
)

type ServiceAccount struct {
	client serviceAccountsClient
	logger logger
	name   string
	email  string
}

func NewServiceAccount(client serviceAccountsClient, logger logger, name, email string) ServiceAccount {
	return ServiceAccount{
		client: client,
		logger: logger,
		name:   name,
		email:  email,
	}
}

func (s ServiceAccount) Delete() error {
	err := s.removeBindings()
	if err != nil {
		return fmt.Errorf("Remove IAM Policy Bindings: %s", err)
	}

	err = s.client.DeleteServiceAccount(s.name)
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

type binding struct {
	ServiceAccount string
	Member         string
	Role           string
}

func (s ServiceAccount) removeBindings() error {
	p, err := s.client.GetProjectIamPolicy()
	if err != nil {
		return fmt.Errorf("Get Project IAM Policy: %s", err)
	}

	for j := len(p.Bindings) - 1; j >= 0; j-- {
		b := p.Bindings[j]

		for i, m := range b.Members {
			if strings.Contains(m, s.email) {
				// Remove this member from the binding
				b.Members = append(b.Members[:i], b.Members[i+1:]...)
			}
		}

		if len(b.Members) == 0 {
			p.Bindings = append(p.Bindings[:j], p.Bindings[j+1:]...)
		} else {
			p.Bindings[j] = b
		}
	}

	_, err = s.client.SetProjectIamPolicy(p)
	if err != nil {
		return fmt.Errorf("Set Project IAM Policy: %s", err)
	}

	return nil
}
