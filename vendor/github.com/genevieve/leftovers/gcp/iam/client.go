package iam

import (
	"fmt"

	gcpiam "google.golang.org/api/iam/v1"
)

type client struct {
	project string
	logger  logger

	service         *gcpiam.Service
	serviceAccounts *gcpiam.ProjectsServiceAccountsService
}

func NewClient(project string, service *gcpiam.Service, logger logger) client {
	return client{
		project:         project,
		logger:          logger,
		service:         service,
		serviceAccounts: service.Projects.ServiceAccounts,
	}
}

func (c client) ListServiceAccounts() (*gcpiam.ListServiceAccountsResponse, error) {
	return c.serviceAccounts.List(fmt.Sprintf("projects/%s", c.project)).Do()
}

func (c client) DeleteServiceAccount(account string) (*gcpiam.Empty, error) {
	return c.serviceAccounts.Delete(account).Do()
}
