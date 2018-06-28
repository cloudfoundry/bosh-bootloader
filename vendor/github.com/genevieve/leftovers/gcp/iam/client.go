package iam

import (
	"fmt"
	"time"

	gcpiam "google.golang.org/api/iam/v1"
)

const PAGE_SIZE = int64(200)

type client struct {
	project string

	service         *gcpiam.Service
	serviceAccounts *gcpiam.ProjectsServiceAccountsService
}

func NewClient(project string, service *gcpiam.Service) client {
	return client{
		project:         project,
		service:         service,
		serviceAccounts: service.Projects.ServiceAccounts,
	}
}

// ListServiceAccounts will loop over every page of results
// and return the full list of service accounts. To prevent
// backend errors from repeated calls, there is a 2s delay.
func (c client) ListServiceAccounts() ([]*gcpiam.ServiceAccount, error) {
	serviceAccounts := []*gcpiam.ServiceAccount{}

	for {
		resp, err := c.serviceAccounts.List(fmt.Sprintf("projects/%s", c.project)).PageSize(PAGE_SIZE).Do()
		if err != nil {
			return serviceAccounts, err
		}

		serviceAccounts = append(serviceAccounts, resp.Accounts...)

		if resp.NextPageToken == "" {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return serviceAccounts, nil
}

func (c client) DeleteServiceAccount(account string) (*gcpiam.Empty, error) {
	return c.serviceAccounts.Delete(account).Do()
}
