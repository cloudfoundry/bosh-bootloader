package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type routesClient interface {
	ListRoutes() ([]*gcpcompute.Route, error)
	DeleteRoute(route string) error
}

type Routes struct {
	client routesClient
	logger logger
}

func NewRoutes(client routesClient, logger logger) Routes {
	return Routes{
		client: client,
		logger: logger,
	}
}

func (r Routes) List(filter string) ([]common.Deletable, error) {
	routes, err := r.client.ListRoutes()
	if err != nil {
		return nil, fmt.Errorf("List Routes: %s", err)
	}

	var resources []common.Deletable
	for _, route := range routes {
		resource := NewRoute(r.client, route.Name)

		if !strings.Contains(route.Name, filter) || strings.Contains(route.Name, "default") {
			continue
		}

		proceed := r.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (r Routes) Type() string {
	return "route"
}
