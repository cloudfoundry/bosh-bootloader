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
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (i InstanceGroup) Name() string {
	return i.name
}

func (i InstanceGroup) Type() string {
	return "Instance Group"
}
