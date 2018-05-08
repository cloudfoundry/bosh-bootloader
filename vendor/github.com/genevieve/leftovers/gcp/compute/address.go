package compute

import "fmt"

type Address struct {
	client      addressesClient
	name        string
	clearerName string
	region      string
	kind        string
}

func NewAddress(client addressesClient, name, region string, users int) Address {
	clearerName := name
	if users > 0 {
		clearerName = fmt.Sprintf("%s (Users:%d)", name, users)
	}

	return Address{
		client:      client,
		name:        name,
		clearerName: clearerName,
		region:      region,
		kind:        "address",
	}
}

func (a Address) Delete() error {
	err := a.client.DeleteAddress(a.region, a.name)
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (a Address) Name() string {
	return a.clearerName
}

func (a Address) Type() string {
	return "Address"
}

func (a Address) Kind() string {
	return a.kind
}
