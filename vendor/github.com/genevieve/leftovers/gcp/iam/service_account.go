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

	toRemove := []binding{}

	for j, b := range p.Bindings {
		for i, m := range b.Members {
			if strings.Contains(m, s.email) {
				toRemove = append(toRemove, binding{
					ServiceAccount: s.email,
					Member:         m,
					Role:           b.Role,
				})

				// Remove this member from the binding
				b.Members = append(b.Members[:i], b.Members[i+1:]...)
			}
		}

		if len(b.Members) == 0 {
			// If there are no more members for the role, remove the whole binding
			p.Bindings = append(p.Bindings[:j], p.Bindings[j+1:]...)
		}
	}

	for _, binding := range toRemove {
		s.logger.Printf("gcloud iam service-accounts remove-iam-policy-binding %s --member %s --role %s\n", binding.ServiceAccount, binding.Member, binding.Role)
	}

	_, err = s.client.SetProjectIamPolicy(p)
	if err != nil {
		return fmt.Errorf("Set Project IAM Policy: %s", err)
	}

	return nil
}
