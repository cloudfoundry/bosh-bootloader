package compute

import "fmt"

type Firewall struct {
	client firewallsClient
	name   string
}

func NewFirewall(client firewallsClient, name string) Firewall {
	return Firewall{
		client: client,
		name:   name,
	}
}

func (f Firewall) Delete() error {
	err := f.client.DeleteFirewall(f.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting firewall %s: %s", f.name, err)
	}

	return nil
}

func (f Firewall) Name() string {
	return f.name
}

func (f Firewall) Type() string {
	return "firewall"
}
