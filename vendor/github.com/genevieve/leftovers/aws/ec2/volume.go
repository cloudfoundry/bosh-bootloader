package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Volume struct {
	client     volumesClient
	id         *string
	identifier string
}

func NewVolume(client volumesClient, id, state *string) Volume {
	return Volume{
		client:     client,
		id:         id,
		identifier: fmt.Sprintf("%s (State:%s)", *id, *state),
	}
}

func (v Volume) Delete() error {
	_, err := v.client.DeleteVolume(&awsec2.DeleteVolumeInput{VolumeId: v.id})
	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVolume.NotFound" {
			return nil
		}
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (v Volume) Name() string {
	return v.identifier
}

func (v Volume) Type() string {
	return "EC2 Volume"
}
