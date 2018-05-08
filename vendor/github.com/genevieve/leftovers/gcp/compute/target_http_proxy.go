package compute

import "fmt"

type TargetHttpProxy struct {
	client targetHttpProxiesClient
	name   string
	kind   string
}

func NewTargetHttpProxy(client targetHttpProxiesClient, name string) TargetHttpProxy {
	return TargetHttpProxy{
		client: client,
		name:   name,
		kind:   "target-http-proxy",
	}
}

func (t TargetHttpProxy) Delete() error {
	err := t.client.DeleteTargetHttpProxy(t.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (t TargetHttpProxy) Name() string {
	return t.name
}

func (t TargetHttpProxy) Type() string {
	return "Target Http Proxy"
}

func (t TargetHttpProxy) Kind() string {
	return t.kind
}
