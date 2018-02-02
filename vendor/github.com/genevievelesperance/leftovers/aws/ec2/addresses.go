package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevievelesperance/leftovers/aws/common"
)

type addressesClient interface {
	DescribeAddresses(*awsec2.DescribeAddressesInput) (*awsec2.DescribeAddressesOutput, error)
	ReleaseAddress(*awsec2.ReleaseAddressInput) (*awsec2.ReleaseAddressOutput, error)
}

type Addresses struct {
	client addressesClient
	logger logger
}

func NewAddresses(client addressesClient, logger logger) Addresses {
	return Addresses{
		client: client,
		logger: logger,
	}
}

func (d Addresses) List(filter string) ([]common.Deletable, error) {
	addresses, err := d.client.DescribeAddresses(&awsec2.DescribeAddressesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing addresses: %s", err)
	}

	var resources []common.Deletable
	for _, a := range addresses.Addresses {
		resource := NewAddress(d.client, a.PublicIp, a.AllocationId)

		if d.inUse(a) {
			continue
		}

		proceed := d.logger.Prompt(fmt.Sprintf("Are you sure you want to release address %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (d Addresses) inUse(a *awsec2.Address) bool {
	return a.InstanceId != nil && *a.InstanceId != ""
}
