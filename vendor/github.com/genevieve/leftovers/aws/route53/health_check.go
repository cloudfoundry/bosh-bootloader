package route53

import (
	"fmt"

	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
)

type HealthCheck struct {
	client     healthChecksClient
	id         *string
	identifier string
}

func NewHealthCheck(client healthChecksClient, id *string) HealthCheck {
	return HealthCheck{
		client:     client,
		id:         id,
		identifier: *id,
	}
}

func (h HealthCheck) Delete() error {
	_, err := h.client.DeleteHealthCheck(&awsroute53.DeleteHealthCheckInput{HealthCheckId: h.id})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (h HealthCheck) Name() string {
	return h.identifier
}

func (h HealthCheck) Type() string {
	return "Route53 Health Check"
}
