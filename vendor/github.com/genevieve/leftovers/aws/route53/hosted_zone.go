package route53

import (
	"fmt"

	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
)

type HostedZone struct {
	client     hostedZonesClient
	id         *string
	identifier string
	recordSets recordSets
}

func NewHostedZone(client hostedZonesClient, id, name *string, recordSets recordSets) HostedZone {
	return HostedZone{
		client:     client,
		id:         id,
		identifier: *name,
		recordSets: recordSets,
	}
}

func (h HostedZone) Delete() error {
	r, err := h.recordSets.Get(h.id)
	if err != nil {
		return fmt.Errorf("Get Record Sets: %s", err)
	}

	err = h.recordSets.Delete(h.id, h.identifier, r)
	if err != nil {
		return fmt.Errorf("Delete Record Sets: %s", err)
	}

	_, err = h.client.DeleteHostedZone(&awsroute53.DeleteHostedZoneInput{Id: h.id})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (h HostedZone) Name() string {
	return h.identifier
}

func (h HostedZone) Type() string {
	return "Route53 Hosted Zone"
}
