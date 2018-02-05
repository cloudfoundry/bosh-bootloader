package compute

import "fmt"

type GlobalForwardingRule struct {
	client globalForwardingRulesClient
	name   string
}

func NewGlobalForwardingRule(client globalForwardingRulesClient, name string) GlobalForwardingRule {
	return GlobalForwardingRule{
		client: client,
		name:   name,
	}
}

func (g GlobalForwardingRule) Delete() error {
	err := g.client.DeleteGlobalForwardingRule(g.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting global forwarding rule %s: %s", g.name, err)
	}

	return nil
}
