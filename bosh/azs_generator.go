package bosh

import "fmt"

type AZsGenerator struct {
	awsNames []string
}

type AZ struct {
	Name            string            `yaml:"name"`
	CloudProperties AZCloudProperties `yaml:"cloud_properties"`
}

type AZCloudProperties struct {
	AvailabilityZone string `yaml:"availability_zone"`
}

func NewAZsGenerator(awsNames ...string) AZsGenerator {
	return AZsGenerator{
		awsNames: awsNames,
	}
}

func (g AZsGenerator) Generate() []AZ {
	AZs := []AZ{}
	for i, awsName := range g.awsNames {
		AZs = append(AZs, AZ{
			Name: fmt.Sprintf("z%d", i+1),
			CloudProperties: AZCloudProperties{
				AvailabilityZone: awsName,
			},
		})
	}
	return AZs
}
