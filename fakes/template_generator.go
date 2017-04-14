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

func (g *TemplateGenerator) Generate(state storage.State) string {
	g.GenerateCall.CallCount++
	g.GenerateCall.Receives.State = state
	return g.GenerateCall.Returns.Template
}
