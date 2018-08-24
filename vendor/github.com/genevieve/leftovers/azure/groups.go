package azure

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/genevieve/leftovers/common"
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

func (g Groups) List(filter string) ([]common.Deletable, error) {
	groups, err := g.client.List("", nil)
	if err != nil {
		return nil, fmt.Errorf("Listing Resource Groups: %s", err)
	}

	var resources []common.Deletable
	for _, group := range *groups.Value {
		r := NewGroup(g.client, group.Name)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := g.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (g Groups) Type() string {
	return "resource-group"
}
