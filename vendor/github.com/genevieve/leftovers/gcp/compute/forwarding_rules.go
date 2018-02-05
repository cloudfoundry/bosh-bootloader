package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type forwardingRulesClient interface {
	ListForwardingRules(region string) (*gcpcompute.ForwardingRuleList, error)
	DeleteForwardingRule(region, rule string) error
}

type ForwardingRules struct {
	client  forwardingRulesClient
	logger  logger
	regions map[string]string
}

func NewForwardingRules(client forwardingRulesClient, logger logger, regions map[string]string) ForwardingRules {
	return ForwardingRules{
		client:  client,
		logger:  logger,
		regions: regions,
	}
}

func (f ForwardingRules) List(filter string) ([]common.Deletable, error) {
	rules := []*gcpcompute.ForwardingRule{}
	for _, region := range f.regions {
		l, err := f.client.ListForwardingRules(region)
		if err != nil {
			return nil, fmt.Errorf("Listing forwarding rules for region %s: %s", region, err)
		}

		rules = append(rules, l.Items...)
	}

	var resources []common.Deletable
	for _, rule := range rules {
		resource := NewForwardingRule(f.client, rule.Name, f.regions[rule.Region])

		if !strings.Contains(rule.Name, filter) {
			continue
		}

		proceed := f.logger.Prompt(fmt.Sprintf("Are you sure you want to delete forwarding rule %s?", rule.Name))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
