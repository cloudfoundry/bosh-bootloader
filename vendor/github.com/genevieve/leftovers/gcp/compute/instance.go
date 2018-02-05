package compute

import "fmt"

type Instance struct {
	client instancesClient
	name   string
	zone   string
}

func NewInstance(client instancesClient, name, zone string) Instance {
	return Instance{
		client: client,
		name:   name,
		zone:   zone,
	}
}

func (i Instance) Delete() error {
	err := i.client.DeleteInstance(i.zone, i.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting instance %s: %s", i.name, err)
	}

	return nil
}
