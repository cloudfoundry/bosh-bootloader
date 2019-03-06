package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type TemplateGenerator struct {
	GenerateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Template string
		}
	}
}

func (t *TemplateGenerator) Generate(state storage.State, withBootstrap bool) string {
	t.GenerateCall.CallCount++
	t.GenerateCall.Receives.State = state
	return t.GenerateCall.Returns.Template
}
