package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type targetPoolsClient interface {
	ListTargetPools(region string) (*gcpcompute.TargetPoolList, error)
	DeleteTargetPool(region string, targetPool string) error
}

type TargetPools struct {
	client  targetPoolsClient
	logger  logger
	regions map[string]string
}

func NewTargetPools(client targetPoolsClient, logger logger, regions map[string]string) TargetPools {
	return TargetPools{
		client:  client,
		logger:  logger,
		regions: regions,
	}
}

func (t TargetPools) List(filter string) ([]common.Deletable, error) {
	pools := []*gcpcompute.TargetPool{}
	for _, region := range t.regions {
		l, err := t.client.ListTargetPools(region)
		if err != nil {
			return nil, fmt.Errorf("List Target Pools for region %s: %s", region, err)
		}

		pools = append(pools, l.Items...)
	}

	var resources []common.Deletable
	for _, pool := range pools {
		resource := NewTargetPool(t.client, pool.Name, t.regions[pool.Region])

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

func (t TargetPools) Type() string {
	return "target-pool"
}
