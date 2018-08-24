package iam

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpiam "google.golang.org/api/iam/v1"
)

type serviceAccountsClient interface {
	ListServiceAccounts() ([]*gcpiam.ServiceAccount, error)
	DeleteServiceAccount(account string) (*gcpiam.Empty, error)
}

type ServiceAccounts struct {
	client serviceAccountsClient
	logger logger
}

func NewServiceAccounts(client serviceAccountsClient, logger logger) ServiceAccounts {
	return ServiceAccounts{
		client: client,
		logger: logger,
	}
}

func (s ServiceAccounts) List(filter string) ([]common.Deletable, error) {
	accounts, err := s.client.ListServiceAccounts()
	if err != nil {
		return nil, fmt.Errorf("List IAM Service Accounts: %s", err)
	}

	var resources []common.Deletable
	for _, account := range accounts {
		resource := NewServiceAccount(s.client, account.Name)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		proceed := s.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (s ServiceAccounts) Type() string {
	return "service-account"
}
