package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
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

func (d Addresses) ListAll(filter string) ([]common.Deletable, error) {
	return d.get(filter)
}

func (d Addresses) List(filter string) ([]common.Deletable, error) {
	resources, err := d.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := d.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (d Addresses) get(filter string) ([]common.Deletable, error) {
	addresses, err := d.client.DescribeAddresses(&awsec2.DescribeAddressesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing EC2 Addresses: %s", err)
	}

	var resources []common.Deletable
	for _, a := range addresses.Addresses {
		resource := NewAddress(d.client, a.PublicIp, a.AllocationId)

		if d.inUse(a) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (d Addresses) inUse(a *awsec2.Address) bool {
	return a.InstanceId != nil && *a.InstanceId != ""
}
