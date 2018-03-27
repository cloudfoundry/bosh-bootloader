package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Volume struct {
	client     volumesClient
	id         *string
	identifier string
	rtype      string
}

func NewVolume(client volumesClient, id *string) Volume {
	return Volume{
		client:     client,
		id:         id,
		identifier: *id,
		rtype:      "EC2 Volume",
	}
}

func (v Volume) Delete() error {
	_, err := v.client.DeleteVolume(&awsec2.DeleteVolumeInput{
		VolumeId: v.id,
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting %s %s: %s", v.rtype, v.identifier, err)
	}

	return nil
}

func (v Volume) Name() string {
	return v.identifier
}

func (v Volume) Type() string {
	return v.rtype
}
