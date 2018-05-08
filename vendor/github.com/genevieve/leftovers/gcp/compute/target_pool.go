package compute

import "fmt"

type TargetPool struct {
	client targetPoolsClient
	name   string
	region string
	kind   string
}

func NewTargetPool(client targetPoolsClient, name, region string) TargetPool {
	return TargetPool{
		client: client,
		name:   name,
		region: region,
		kind:   "target-pool",
	}
}

func (t TargetPool) Delete() error {
	err := t.client.DeleteTargetPool(t.region, t.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (t TargetPool) Name() string {
	return t.name
}

func (t TargetPool) Type() string {
	return "Target Pool"
}

func (t TargetPool) Kind() string {
	return t.kind
}
