package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Address struct {
	client       addressesClient
	publicIp     *string
	allocationId *string
	identifier   string
}

func NewAddress(client addressesClient, publicIp, allocationId *string) Address {
	return Address{
		client:       client,
		publicIp:     publicIp,
		allocationId: allocationId,
		identifier:   *publicIp,
	}
}

func (a Address) Delete() error {
	_, err := a.client.ReleaseAddress(&awsec2.ReleaseAddressInput{
		AllocationId: a.allocationId,
	})

	if err != nil {
		return fmt.Errorf("FAILED releasing address %s: %s", a.identifier, err)
	}

	return nil
}

func (a Address) Name() string {
	return a.identifier
}
