package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type targetHttpProxiesClient interface {
	ListTargetHttpProxies() (*gcpcompute.TargetHttpProxyList, error)
	DeleteTargetHttpProxy(targetHttpProxy string) error
}

type TargetHttpProxies struct {
	client targetHttpProxiesClient
	logger logger
}

func NewTargetHttpProxies(client targetHttpProxiesClient, logger logger) TargetHttpProxies {
	return TargetHttpProxies{
		client: client,
		logger: logger,
	}
}

func (t TargetHttpProxies) List(filter string) ([]common.Deletable, error) {
	targetHttpProxies, err := t.client.ListTargetHttpProxies()
	if err != nil {
		return nil, fmt.Errorf("Listing target http proxies: %s", err)
	}

	var resources []common.Deletable
	for _, targetHttpProxy := range targetHttpProxies.Items {
		resource := NewTargetHttpProxy(t.client, targetHttpProxy.Name)

		if !strings.Contains(resource.name, filter) {
			continue
		}

		proceed := t.logger.Prompt(fmt.Sprintf("Are you sure you want to delete target http proxy %s?", resource.name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
