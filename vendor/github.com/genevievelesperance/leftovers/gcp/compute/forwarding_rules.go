package compute

import (
	"fmt"
	"strings"
	"sync"

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

func (f ForwardingRules) List(filter string) (map[string]string, error) {
	rules := []*gcpcompute.ForwardingRule{}
	for _, region := range f.regions {
		l, err := f.client.ListForwardingRules(region)
		if err != nil {
			return nil, fmt.Errorf("Listing forwarding rules for region %s: %s", region, err)
		}

		rules = append(rules, l.Items...)
	}

	delete := map[string]string{}
	for _, rule := range rules {
		if !strings.Contains(rule.Name, filter) {
			continue
		}

		proceed := f.logger.Prompt(fmt.Sprintf("Are you sure you want to delete forwarding rule %s?", rule.Name))
		if !proceed {
			continue
		}

		delete[rule.Name] = f.regions[rule.Region]
	}

	return delete, nil
}

func (f ForwardingRules) Delete(forwardingRules map[string]string) {
	var wg sync.WaitGroup

	for name, region := range forwardingRules {
		wg.Add(1)

		go func(name, region string) {
			err := f.client.DeleteForwardingRule(region, name)

			if err != nil {
				f.logger.Printf("ERROR deleting forwarding rule %s: %s\n", name, err)
			} else {
				f.logger.Printf("SUCCESS deleting forwarding rule %s\n", name)
			}

			wg.Done()
		}(name, region)
	}

	wg.Wait()
}
