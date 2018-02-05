package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type targetHttpsProxiesClient interface {
	ListTargetHttpsProxies() (*gcpcompute.TargetHttpsProxyList, error)
	DeleteTargetHttpsProxy(targetHttpsProxy string) error
}

type TargetHttpsProxies struct {
	client targetHttpsProxiesClient
	logger logger
}

func NewTargetHttpsProxies(client targetHttpsProxiesClient, logger logger) TargetHttpsProxies {
	return TargetHttpsProxies{
		client: client,
		logger: logger,
	}
}

func (t TargetHttpsProxies) List(filter string) ([]common.Deletable, error) {
	targetHttpsProxies, err := t.client.ListTargetHttpsProxies()
	if err != nil {
		return nil, fmt.Errorf("Listing target https proxies: %s", err)
	}

	var resources []common.Deletable
	for _, targetHttpsProxy := range targetHttpsProxies.Items {
		resource := NewTargetHttpsProxy(t.client, targetHttpsProxy.Name)

		if !strings.Contains(resource.name, filter) {
			continue
		}

		proceed := t.logger.Prompt(fmt.Sprintf("Are you sure you want to delete target https proxy %s?", resource.name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
