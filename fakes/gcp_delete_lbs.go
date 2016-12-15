package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPDeleteLBs struct {
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}

		Returns struct {
			Error error
		}
	}
}

func (g *GCPDeleteLBs) Execute(state storage.State) error {
	g.ExecuteCall.CallCount++
	g.ExecuteCall.Receives.State = state
	return g.ExecuteCall.Returns.Error
}
