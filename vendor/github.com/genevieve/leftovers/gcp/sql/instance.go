package sql

import "fmt"

type Instance struct {
	client instancesClient
	name   string
	kind   string
}

func NewInstance(client instancesClient, name string) Instance {
	return Instance{
		client: client,
		name:   name,
		kind:   "sql-instance",
	}
}
func (i Instance) Delete() error {
	err := i.client.DeleteInstance(i.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (i Instance) Name() string {
	return i.name
}

func (i Instance) Type() string {
	return "SQL Instance"
}

func (i Instance) Kind() string {
	return i.kind
}
