package compute

import "fmt"

type Route struct {
	client routesClient
	name   string
}

func NewRoute(client routesClient, name string) Route {
	return Route{
		client: client,
		name:   name,
	}
}

func (r Route) Delete() error {
	err := r.client.DeleteRoute(r.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (r Route) Name() string {
	return r.name
}

func (r Route) Type() string {
	return "Route"
}
