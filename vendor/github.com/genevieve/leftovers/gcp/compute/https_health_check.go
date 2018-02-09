package compute

import "fmt"

type HttpsHealthCheck struct {
	client httpsHealthChecksClient
	name   string
}

func NewHttpsHealthCheck(client httpsHealthChecksClient, name string) HttpsHealthCheck {
	return HttpsHealthCheck{
		client: client,
		name:   name,
	}
}

func (h HttpsHealthCheck) Delete() error {
	err := h.client.DeleteHttpsHealthCheck(h.name)

	if err != nil {
		return fmt.Errorf("ERROR deleting https health check %s: %s", h.name, err)
	}

	return nil
}

func (h HttpsHealthCheck) Name() string {
	return h.name
}
