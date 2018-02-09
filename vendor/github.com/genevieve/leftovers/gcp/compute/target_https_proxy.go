package compute

import "fmt"

type TargetHttpsProxy struct {
	client targetHttpsProxiesClient
	name   string
}

func NewTargetHttpsProxy(client targetHttpsProxiesClient, name string) TargetHttpsProxy {
	return TargetHttpsProxy{
		client: client,
		name:   name,
	}
}

func (t TargetHttpsProxy) Delete() error {
	err := t.client.DeleteTargetHttpsProxy(t.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting target https proxy %s: %s", t.name, err)
	}

	return nil
}

func (t TargetHttpsProxy) Name() string {
	return t.name
}
