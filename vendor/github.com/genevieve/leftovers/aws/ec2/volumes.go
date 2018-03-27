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

func (v Volumes) ListAll(filter string) ([]common.Deletable, error) {
	return v.get(filter)
}

func (v Volumes) List(filter string) ([]common.Deletable, error) {
	resources, err := v.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := v.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (v Volumes) get(filter string) ([]common.Deletable, error) {
	output, err := v.client.DescribeVolumes(&awsec2.DescribeVolumesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing volumes: %s", err)
	}

	var resources []common.Deletable
	for _, volume := range output.Volumes {
		resource := NewVolume(v.client, volume.VolumeId)

		if *volume.State != "available" {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
