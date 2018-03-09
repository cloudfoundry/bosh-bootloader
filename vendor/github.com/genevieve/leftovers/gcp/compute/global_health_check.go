package compute

import "fmt"

type GlobalHealthCheck struct {
	client globalHealthChecksClient
	name   string
}

func NewGlobalHealthCheck(client globalHealthChecksClient, name string) GlobalHealthCheck {
	return GlobalHealthCheck{
		client: client,
		name:   name,
	}
}

func (g GlobalHealthCheck) Delete() error {
	err := g.client.DeleteGlobalHealthCheck(g.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting global health check %s: %s", g.name, err)
	}

	return nil
}

func (g GlobalHealthCheck) Name() string {
	return g.name
}

func (g GlobalHealthCheck) Type() string {
	return "global health check"
}
