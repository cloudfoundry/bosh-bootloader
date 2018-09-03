package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Address struct {
	client       addressesClient
	publicIp     *string
	allocationId *string
	identifier   string
	rtype        string
}

func NewAddress(client addressesClient, publicIp, allocationId *string, tags []*awsec2.Tag) Address {
	identifier := *publicIp

	var extra []string
	for _, t := range tags {
		extra = append(extra, fmt.Sprintf("%s:%s", *t.Key, *t.Value))
	}

	if len(extra) > 0 {
		identifier = fmt.Sprintf("%s (%s)", identifier, strings.Join(extra, ","))
	}

	return Address{
		client:       client,
		publicIp:     publicIp,
		allocationId: allocationId,
		identifier:   identifier,
		rtype:        "EC2 Address",
	}
}

func (a Address) Delete() error {
	_, err := a.client.ReleaseAddress(&awsec2.ReleaseAddressInput{AllocationId: a.allocationId})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidAllocationID.NotFound" {
				return nil
			}
		}
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (a Address) Name() string {
	return a.identifier
}

func (a Address) Type() string {
	return a.rtype
}
