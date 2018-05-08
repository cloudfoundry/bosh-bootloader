package compute

import "fmt"

type GlobalForwardingRule struct {
	client globalForwardingRulesClient
	name   string
	kind   string
}

func NewGlobalForwardingRule(client globalForwardingRulesClient, name string) GlobalForwardingRule {
	return GlobalForwardingRule{
		client: client,
		name:   name,
		kind:   "global-forwarding-rule",
	}
}

func (g GlobalForwardingRule) Delete() error {
	err := g.client.DeleteGlobalForwardingRule(g.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (g GlobalForwardingRule) Name() string {
	return g.name
}

func (g GlobalForwardingRule) Type() string {
	return "Global Forwarding Rule"
}

func (g GlobalForwardingRule) Kind() string {
	return g.kind
}
