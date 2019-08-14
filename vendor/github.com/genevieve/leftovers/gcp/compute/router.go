package compute

import "fmt"

type Router struct {
	client routersClient
	name   string
	region string
}

func NewRouter(client routersClient, name, region string) Router {
	return Router{
		client: client,
		name:   name,
		region: region,
	}
}

func (r Router) Delete() error {
	err := r.client.DeleteRouter(r.region, r.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (r Router) Type() string {
	return "Router"
}

func (r Router) Name() string {
	return r.name
}
