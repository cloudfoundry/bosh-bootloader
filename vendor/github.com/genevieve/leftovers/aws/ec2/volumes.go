package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
)

type volumesClient interface {
	DescribeVolumes(*awsec2.DescribeVolumesInput) (*awsec2.DescribeVolumesOutput, error)
	DeleteVolume(*awsec2.DeleteVolumeInput) (*awsec2.DeleteVolumeOutput, error)
}

type Volumes struct {
	client volumesClient
	logger logger
}

func NewVolumes(client volumesClient, logger logger) Volumes {
	return Volumes{
		client: client,
		logger: logger,
	}
}

func (v Volumes) List(filter string) ([]common.Deletable, error) {
	output, err := v.client.DescribeVolumes(&awsec2.DescribeVolumesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing volumes: %s", err)
	}

	var volumes []common.Deletable
	for _, volume := range output.Volumes {
		if *volume.State != "available" {
			continue
		}

		proceed := v.logger.Prompt(fmt.Sprintf("Are you sure you want to delete volume %s?", *volume.VolumeId))
		if !proceed {
			continue
		}

		volumes = append(volumes, NewVolume(v.client, volume.VolumeId))
	}

	return volumes, nil
}
