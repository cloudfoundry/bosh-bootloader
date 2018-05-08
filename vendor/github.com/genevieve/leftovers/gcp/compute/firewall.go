package compute

import "fmt"

type Firewall struct {
	client firewallsClient
	name   string
	kind   string
}

func NewFirewall(client firewallsClient, name string) Firewall {
	return Firewall{
		client: client,
		name:   name,
		kind:   "firewall",
	}
}

func (f Firewall) Delete() error {
	err := f.client.DeleteFirewall(f.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (f Firewall) Name() string {
	return f.name
}

func (f Firewall) Type() string {
	return "Firewall"
}

func (f Firewall) Kind() string {
	return f.kind
}
