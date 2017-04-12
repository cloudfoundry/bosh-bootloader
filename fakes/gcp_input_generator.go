package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPInputGenerator struct {
	GenerateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Inputs map[string]string
			Error  error
		}
	}
}

func (g *GCPInputGenerator) Generate(state storage.State) (map[string]string, error) {
	g.GenerateCall.CallCount++
	g.GenerateCall.Receives.State = state
	return g.GenerateCall.Returns.Inputs, g.GenerateCall.Returns.Error
}
