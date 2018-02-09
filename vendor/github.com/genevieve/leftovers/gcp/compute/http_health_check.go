package compute

import "fmt"

type HttpHealthCheck struct {
	client httpHealthChecksClient
	name   string
}

func NewHttpHealthCheck(client httpHealthChecksClient, name string) HttpHealthCheck {
	return HttpHealthCheck{
		client: client,
		name:   name,
	}
}

func (h HttpHealthCheck) Delete() error {
	err := h.client.DeleteHttpHealthCheck(h.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting http health check %s: %s", h.name, err)
	}

	return nil
}

func (h HttpHealthCheck) Name() string {
	return h.name
}
