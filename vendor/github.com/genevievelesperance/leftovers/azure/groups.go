package azure

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
)

type groupsClient interface {
	List(query string, top *int32) (resources.GroupListResult, error)
	Delete(name string, channel <-chan struct{}) (<-chan autorest.Response, <-chan error)
}

type Groups struct {
	client groupsClient
	logger logger
}

func NewGroups(client groupsClient, logger logger) Groups {
	return Groups{
		client: client,
		logger: logger,
	}
}

func (g Groups) List(filter string) ([]string, error) {
	delete := []string{}

	groups, err := g.client.List("", nil)
	if err != nil {
		return delete, fmt.Errorf("Listing resource groups: %s", err)
	}

	for _, group := range *groups.Value {
		n := *group.Name

		if !strings.Contains(n, filter) {
			continue
		}

		proceed := g.logger.Prompt(fmt.Sprintf("Are you sure you want to delete resource group %s?", n))
		if !proceed {
			continue
		}

		delete = append(delete, n)
	}

	return delete, nil
}

func (g Groups) Delete(groups []string) error {
	for _, name := range groups {
		_, errChan := g.client.Delete(name, nil)

		err := <-errChan

		if err == nil {
			g.logger.Printf("SUCCESS deleting resource group %s\n", name)
		} else {
			g.logger.Printf("ERROR deleting resource group %s: %s\n", name, err)
		}
	}

	return nil
}
