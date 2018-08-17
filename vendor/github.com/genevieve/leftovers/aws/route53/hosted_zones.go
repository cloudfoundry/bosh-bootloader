package route53

import (
	"fmt"
	"strings"

	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
	"github.com/genevieve/leftovers/common"
)

type hostedZonesClient interface {
	ListHostedZones(*awsroute53.ListHostedZonesInput) (*awsroute53.ListHostedZonesOutput, error)
	DeleteHostedZone(*awsroute53.DeleteHostedZoneInput) (*awsroute53.DeleteHostedZoneOutput, error)

	ListResourceRecordSets(*awsroute53.ListResourceRecordSetsInput) (*awsroute53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(*awsroute53.ChangeResourceRecordSetsInput) (*awsroute53.ChangeResourceRecordSetsOutput, error)
}

type HostedZones struct {
	client hostedZonesClient
	logger logger
}

func NewHostedZones(client hostedZonesClient, logger logger) HostedZones {
	return HostedZones{
		client: client,
		logger: logger,
	}
}

func (z HostedZones) List(filter string) ([]common.Deletable, error) {
	zones, err := z.client.ListHostedZones(&awsroute53.ListHostedZonesInput{})
	if err != nil {
		return nil, fmt.Errorf("List Route53 Hosted Zones: %s", err)
	}

	var resources []common.Deletable
	for _, zone := range zones.HostedZones {
		r := NewHostedZone(z.client, zone.Id, zone.Name)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := z.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (z HostedZones) Type() string {
	return "route53-hosted-zone"
}
