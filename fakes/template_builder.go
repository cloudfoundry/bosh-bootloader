package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

type TemplateBuilder struct {
	BuildCall struct {
		Receives struct {
			KeyPairName string
		}
		Returns struct {
			Template templates.Template
		}
	}
}

func (b *TemplateBuilder) Build(keyPairName string) templates.Template {
	b.BuildCall.Receives.KeyPairName = keyPairName

	return b.BuildCall.Returns.Template
}
