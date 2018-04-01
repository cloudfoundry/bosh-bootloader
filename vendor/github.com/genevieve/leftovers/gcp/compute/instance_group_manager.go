package compute

import "fmt"

type InstanceGroupManager struct {
	client instanceGroupManagersClient
	name   string
	zone   string
}

func NewInstanceGroupManager(client instanceGroupManagersClient, name, zone string) InstanceGroupManager {
	return InstanceGroupManager{
		client: client,
		name:   name,
		zone:   zone,
	}
}

func (i InstanceGroupManager) Delete() error {
	err := i.client.DeleteInstanceGroupManager(i.zone, i.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (i InstanceGroupManager) Name() string {
	return i.name
}

func (i InstanceGroupManager) Type() string {
	return "Instance Group Manager"
}
