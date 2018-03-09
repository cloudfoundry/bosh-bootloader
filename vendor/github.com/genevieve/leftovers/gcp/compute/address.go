package compute

import "fmt"

type Address struct {
	client addressesClient
	name   string
	region string
}

func NewAddress(client addressesClient, name, region string) Address {
	return Address{
		client: client,
		name:   name,
		region: region,
	}
}

func (a Address) Delete() error {
	err := a.client.DeleteAddress(a.region, a.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting address %s: %s", a.name, err)
	}

	return nil
}

func (a Address) Name() string {
	return a.name
}

func (a Address) Type() string {
	return "address"
}
