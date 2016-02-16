package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type TemplateBuilder struct {
	BuildCall struct {
		Returns struct {
			Template cloudformation.Template
		}
	}
}

func (b TemplateBuilder) Build() cloudformation.Template {
	return b.BuildCall.Returns.Template
}
