package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type TemplateBuilder struct {
	BuildCall struct {
		Receives struct {
			KeyPairName string
		}
		Returns struct {
			Template cloudformation.Template
		}
	}
}

func (b *TemplateBuilder) Build(keyPairName string) cloudformation.Template {
	b.BuildCall.Receives.KeyPairName = keyPairName

	return b.BuildCall.Returns.Template
}
