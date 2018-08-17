package sql

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpsql "google.golang.org/api/sqladmin/v1beta4"
)

type instancesClient interface {
	ListInstances() (*gcpsql.InstancesListResponse, error)
	DeleteInstance(user string) error
}

type Instances struct {
	client instancesClient
	logger logger
}

func NewInstances(client instancesClient, logger logger) Instances {
	return Instances{
		client: client,
		logger: logger,
	}
}

func (i Instances) List(filter string) ([]common.Deletable, error) {
	instances, err := i.client.ListInstances()
	if err != nil {
		return nil, fmt.Errorf("List SQL Instances: %s", err)
	}

	var resources []common.Deletable
	for _, instance := range instances.Items {
		resource := NewInstance(i.client, instance.Name)

		if !strings.Contains(resource.name, filter) {
			continue
		}

		proceed := i.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (i Instances) Type() string {
	return "sql-instance"
}
