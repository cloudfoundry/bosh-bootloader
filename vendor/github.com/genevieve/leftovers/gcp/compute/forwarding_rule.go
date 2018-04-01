package compute

import "fmt"

type ForwardingRule struct {
	client forwardingRulesClient
	name   string
	region string
}

func NewForwardingRule(client forwardingRulesClient, name, region string) ForwardingRule {
	return ForwardingRule{
		client: client,
		name:   name,
		region: region,
	}
}

func (f ForwardingRule) Delete() error {
	err := f.client.DeleteForwardingRule(f.region, f.name)
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (f ForwardingRule) Name() string {
	return f.name
}

func (f ForwardingRule) Type() string {
	return "Forwarding Rule"
}
