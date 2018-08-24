package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
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
		return nil, fmt.Errorf("List Target Https Proxies: %s", err)
	}

	var resources []common.Deletable
	for _, targetHttpsProxy := range targetHttpsProxies.Items {
		resource := NewTargetHttpsProxy(t.client, targetHttpsProxy.Name)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		proceed := t.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (t TargetHttpsProxies) Type() string {
	return "target-https-proxy"
}
