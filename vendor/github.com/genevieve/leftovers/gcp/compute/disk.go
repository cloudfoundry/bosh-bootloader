package compute

import "fmt"

type Disk struct {
	client disksClient
	name   string
	zone   string
	kind   string
}

func NewDisk(client disksClient, name, zone string) Disk {
	return Disk{
		client: client,
		name:   name,
		zone:   zone,
		kind:   "disk",
	}
}

func (d Disk) Delete() error {
	err := d.client.DeleteDisk(d.zone, d.name)
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (d Disk) Name() string {
	return d.name
}

func (d Disk) Type() string {
	return "Disk"
}

func (d Disk) Kind() string {
	return d.kind
}
