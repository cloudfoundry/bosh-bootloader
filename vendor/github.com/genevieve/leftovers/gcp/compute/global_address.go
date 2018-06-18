package compute

import "fmt"

type GlobalAddress struct {
	client globalAddressesClient
	name   string
}

func NewGlobalAddress(client globalAddressesClient, name string) GlobalAddress {
	return GlobalAddress{
		client: client,
		name:   name,
	}
}

func (g GlobalAddress) Delete() error {
	err := g.client.DeleteGlobalAddress(g.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (g GlobalAddress) Name() string {
	return g.name
}

func (g GlobalAddress) Type() string {
	return "Global Address"
}
