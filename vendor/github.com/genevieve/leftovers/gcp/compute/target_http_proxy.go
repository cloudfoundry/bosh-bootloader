package compute

import "fmt"

type TargetHttpProxy struct {
	client targetHttpProxiesClient
	name   string
}

func NewTargetHttpProxy(client targetHttpProxiesClient, name string) TargetHttpProxy {
	return TargetHttpProxy{
		client: client,
		name:   name,
	}
}

func (t TargetHttpProxy) Delete() error {
	err := t.client.DeleteTargetHttpProxy(t.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting target http proxy %s: %s", t.name, err)
	}

	return nil
}
