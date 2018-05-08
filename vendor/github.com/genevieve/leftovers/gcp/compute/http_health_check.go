package compute

import "fmt"

type HttpHealthCheck struct {
	client httpHealthChecksClient
	name   string
	kind   string
}

func NewHttpHealthCheck(client httpHealthChecksClient, name string) HttpHealthCheck {
	return HttpHealthCheck{
		client: client,
		name:   name,
		kind:   "http-health-check",
	}
}

func (h HttpHealthCheck) Delete() error {
	err := h.client.DeleteHttpHealthCheck(h.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (h HttpHealthCheck) Name() string {
	return h.name
}

func (h HttpHealthCheck) Type() string {
	return "Http Health Check"
}

func (h HttpHealthCheck) Kind() string {
	return h.kind
}
