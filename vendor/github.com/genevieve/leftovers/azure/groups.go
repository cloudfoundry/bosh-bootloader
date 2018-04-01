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

func (g Groups) ListOnly(filter string) ([]Deletable, error) {
	return g.get(filter)
}

func (g Groups) List(filter string) ([]Deletable, error) {
	resources, err := g.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []Deletable
	for _, r := range resources {
		proceed := g.logger.Prompt(fmt.Sprintf("Are you sure you want to delete %s %s?", r.Type(), r.Name()))
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (g Groups) get(filter string) ([]Deletable, error) {
	groups, err := g.client.List("", nil)
	if err != nil {
		return nil, fmt.Errorf("Listing Resource Groups: %s", err)
	}

	var resources []Deletable
	for _, group := range *groups.Value {
		resource := NewGroup(g.client, group.Name)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
