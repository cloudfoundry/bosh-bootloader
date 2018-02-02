package compute

import (
	"fmt"
	"strings"
	"sync"

	gcpcompute "google.golang.org/api/compute/v1"
)

type globalForwardingRulesClient interface {
	ListGlobalForwardingRules() (*gcpcompute.ForwardingRuleList, error)
	DeleteGlobalForwardingRule(rule string) error
}

type GlobalForwardingRules struct {
	client globalForwardingRulesClient
	logger logger
}

func NewGlobalForwardingRules(client globalForwardingRulesClient, logger logger) GlobalForwardingRules {
	return GlobalForwardingRules{
		client: client,
		logger: logger,
	}
}

func (g GlobalForwardingRules) List(filter string) (map[string]string, error) {
	rules, err := g.client.ListGlobalForwardingRules()
	if err != nil {
		return nil, fmt.Errorf("Listing global forwarding rules: %s", err)
	}

	delete := map[string]string{}
	for _, rule := range rules.Items {
		if !strings.Contains(rule.Name, filter) {
			continue
		}

		proceed := g.logger.Prompt(fmt.Sprintf("Are you sure you want to delete global forwarding rule %s?", rule.Name))
		if !proceed {
			continue
		}

		delete[rule.Name] = ""
	}

	return delete, nil
}

func (g GlobalForwardingRules) Delete(globalForwardingRules map[string]string) {
	var wg sync.WaitGroup

	for name, _ := range globalForwardingRules {
		wg.Add(1)

		go func(name string) {
			err := g.client.DeleteGlobalForwardingRule(name)

			if err != nil {
				g.logger.Printf("ERROR deleting global forwarding rule %s: %s\n", name, err)
			} else {
				g.logger.Printf("SUCCESS deleting global forwarding rule %s\n", name)
			}

			wg.Done()
		}(name)
	}

	wg.Wait()
}
