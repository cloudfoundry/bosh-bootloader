package compute

import "fmt"

type TargetHttpsProxy struct {
	client targetHttpsProxiesClient
	name   string
	kind   string
}

func NewTargetHttpsProxy(client targetHttpsProxiesClient, name string) TargetHttpsProxy {
	return TargetHttpsProxy{
		client: client,
		name:   name,
		kind:   "target-https-proxy",
	}
}

func (t TargetHttpsProxy) Delete() error {
	err := t.client.DeleteTargetHttpsProxy(t.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (t TargetHttpsProxy) Name() string {
	return t.name
}

func (t TargetHttpsProxy) Type() string {
	return "Target Https Proxy"
}

func (t TargetHttpsProxy) Kind() string {
	return t.kind
}
