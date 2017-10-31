package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type InputGenerator struct {
	GenerateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Inputs map[string]interface{}
			Error  error
		}
	}
}

func (i *InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	i.GenerateCall.CallCount++
	i.GenerateCall.Receives.State = state
	return i.GenerateCall.Returns.Inputs, i.GenerateCall.Returns.Error
}
