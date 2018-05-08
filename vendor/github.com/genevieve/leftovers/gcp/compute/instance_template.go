package compute

import "fmt"

type InstanceTemplate struct {
	client instanceTemplatesClient
	name   string
	kind   string
}

func NewInstanceTemplate(client instanceTemplatesClient, name string) InstanceTemplate {
	return InstanceTemplate{
		client: client,
		name:   name,
		kind:   "instance-template",
	}
}

func (i InstanceTemplate) Delete() error {
	err := i.client.DeleteInstanceTemplate(i.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (i InstanceTemplate) Name() string {
	return i.name
}

func (i InstanceTemplate) Type() string {
	return "Instance Template"
}

func (i InstanceTemplate) Kind() string {
	return i.kind
}
