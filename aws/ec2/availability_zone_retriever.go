package ec2

import (
	"errors"

	goaws "github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type AvailabilityZoneRetriever struct {
	client Client
}

func NewAvailabilityZoneRetriever(client Client) AvailabilityZoneRetriever {
	return AvailabilityZoneRetriever{
		client: client,
	}
}

func (r AvailabilityZoneRetriever) Retrieve(region string) ([]string, error) {
	output, err := r.client.DescribeAvailabilityZones(&awsec2.DescribeAvailabilityZonesInput{
		Filters: []*awsec2.Filter{{
			Name:   goaws.String("region-name"),
			Values: []*string{goaws.String(region)},
		}},
	})
	if err != nil {
		return []string{}, err
	}

	azList := []string{}
	for _, az := range output.AvailabilityZones {
		if az == nil {
			return []string{}, errors.New("aws returned nil availability zone")
		}
		if az.ZoneName == nil {
			return []string{}, errors.New("aws returned availability zone with nil zone name")
		}

		azList = append(azList, *az.ZoneName)
	}

	return azList, nil
}
