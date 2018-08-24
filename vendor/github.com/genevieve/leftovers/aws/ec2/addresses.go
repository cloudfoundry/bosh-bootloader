package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/common"
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
		return nil, fmt.Errorf("Describing EC2 Addresses: %s", err)
	}

	var resources []common.Deletable
	for _, a := range addresses.Addresses {
		if a.InstanceId != nil && *a.InstanceId != "" {
			continue
		}

		r := NewAddress(d.client, a.PublicIp, a.AllocationId)

		proceed := d.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (d Addresses) Type() string {
	return "ec2-address"
}
