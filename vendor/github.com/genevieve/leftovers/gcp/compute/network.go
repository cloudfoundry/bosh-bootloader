package compute

import "fmt"

type Network struct {
	client networksClient
	name   string
}

func NewNetwork(client networksClient, name string) Network {
	return Network{
		client: client,
		name:   name,
	}
}

func (n Network) Delete() error {
	err := n.client.DeleteNetwork(n.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (n Network) Name() string {
	return n.name
}

func (n Network) Type() string {
	return "Network"
}
