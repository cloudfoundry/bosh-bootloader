package iam

import (
	"fmt"
	"time"

	gcpcrm "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	gcpiam "google.golang.org/api/iam/v1"
)

var mutexKV = NewMutexKV()

type client struct {
	project string

	serviceAccounts *gcpiam.ProjectsServiceAccountsService
	projects        *gcpcrm.ProjectsService
}

func NewClient(project string, iamService *gcpiam.Service, crmService *gcpcrm.Service) client {
	return client{
		project:         project,
		serviceAccounts: iamService.Projects.ServiceAccounts,
		projects:        crmService.Projects,
	}
}

func (c client) GetProjectIamPolicy() (*gcpcrm.Policy, error) {
	return c.projects.GetIamPolicy(c.project, &gcpcrm.GetIamPolicyRequest{}).Do()
}

func (c client) SetProjectIamPolicy(p *gcpcrm.Policy) (*gcpcrm.Policy, error) {
	return c.projects.SetIamPolicy(c.project, &gcpcrm.SetIamPolicyRequest{Policy: p}).Do()
}

// ListServiceAccounts will loop over every page of results
// and return the full list of service accounts. To prevent
// backend errors from repeated calls, there is a 2s delay.
func (c client) ListServiceAccounts() ([]*gcpiam.ServiceAccount, error) {
	var token string
	serviceAccounts := []*gcpiam.ServiceAccount{}

	for {
		resp, err := c.serviceAccounts.List(fmt.Sprintf("projects/%s", c.project)).PageToken(token).Do()
		if err != nil {
			return []*gcpiam.ServiceAccount{}, err
		}

		serviceAccounts = append(serviceAccounts, resp.Accounts...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return serviceAccounts, nil
}

func (c client) DeleteServiceAccount(account string) error {
	_, err := c.serviceAccounts.Delete(account).Do()
	if err != nil {
		gerr, ok := err.(*googleapi.Error)
		if ok && gerr != nil && gerr.Code == 404 {
			return nil
		}

		return err
	}
	return nil
}
