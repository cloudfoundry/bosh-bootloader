package gcp

import "fmt"

type AZsGenerator struct {
	gcpNames []string
}

type AZ struct {
	Name            string            `yaml:"name"`
	CloudProperties AZCloudProperties `yaml:"cloud_properties"`
}

type AZCloudProperties struct {
	Zone string `yaml:"zone"`
}

func NewAZsGenerator(gcpNames ...string) AZsGenerator {
	return AZsGenerator{
		gcpNames: gcpNames,
	}
}

func (g AZsGenerator) Generate() []AZ {
	AZs := []AZ{}
	for i, gcpName := range g.gcpNames {
		AZs = append(AZs, AZ{
			Name: fmt.Sprintf("z%d", i+1),
			CloudProperties: AZCloudProperties{
				Zone: gcpName,
			},
		})
	}
	return AZs
}
