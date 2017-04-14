package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type OutputGenerator struct {
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

func (o *OutputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	o.GenerateCall.CallCount++
	o.GenerateCall.Receives.State = state
	return o.GenerateCall.Returns.Outputs, o.GenerateCall.Returns.Error
}
