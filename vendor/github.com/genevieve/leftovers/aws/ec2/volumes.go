package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/common"
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
	output, err := v.client.DescribeVolumes(&awsec2.DescribeVolumesInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("status"),
			Values: []*string{aws.String("available")},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("Describe EC2 Volumes: %s", err)
	}

	var resources []common.Deletable
	for _, volume := range output.Volumes {
		r := NewVolume(v.client, volume.VolumeId, volume.State)

		proceed := v.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (v Volumes) Type() string {
	return "ec2-volume"
}
