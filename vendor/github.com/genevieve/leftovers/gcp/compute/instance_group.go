package compute

import "fmt"

type InstanceGroup struct {
	client instanceGroupsClient
	name   string
	zone   string
}

func NewInstanceGroup(client instanceGroupsClient, name, zone string) InstanceGroup {
	return InstanceGroup{
		client: client,
		name:   name,
		zone:   zone,
	}
}

func (i InstanceGroup) Delete() error {
	err := i.client.DeleteInstanceGroup(i.zone, i.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting instance group %s: %s", i.name, err)
	}

	return nil
}
