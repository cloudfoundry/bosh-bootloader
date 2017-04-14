package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPOutputGenerator struct {
	GenerateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Outputs map[string]interface{}
			Error   error
		}
	}
}

func (g *GCPOutputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	g.GenerateCall.CallCount++
	g.GenerateCall.Receives.State = state
	return g.GenerateCall.Returns.Outputs, g.GenerateCall.Returns.Error
}
