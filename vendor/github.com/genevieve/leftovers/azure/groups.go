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

func (g Groups) List(filter string) ([]Deletable, error) {
	groups, err := g.client.List("", nil)
	if err != nil {
		return nil, fmt.Errorf("Listing resource groups: %s", err)
	}

	var resources []Deletable
	for _, group := range *groups.Value {
		resource := NewGroup(g.client, group.Name)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := g.logger.Prompt(fmt.Sprintf("Are you sure you want to delete resource group %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
