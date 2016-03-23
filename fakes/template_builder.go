package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

type TemplateBuilder struct {
	BuildCall struct {
		Receives struct {
			KeyPairName string
			NumberOfAZs int
		}
		Returns struct {
			Template templates.Template
		}
	}
}

func (b *TemplateBuilder) Build(keyPairName string, numberOfAvailabilityZones int) templates.Template {
	b.BuildCall.Receives.KeyPairName = keyPairName
	b.BuildCall.Receives.NumberOfAZs = numberOfAvailabilityZones

	return b.BuildCall.Returns.Template
}
