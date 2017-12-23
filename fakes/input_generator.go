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
	CredentialsCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Credentials map[string]string
		}
	}
}

func (i *InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	i.GenerateCall.CallCount++
	i.GenerateCall.Receives.State = state
	return i.GenerateCall.Returns.Inputs, i.GenerateCall.Returns.Error
}

func (i *InputGenerator) Credentials(state storage.State) map[string]string {
	i.CredentialsCall.CallCount++
	i.CredentialsCall.Receives.State = state
	return i.CredentialsCall.Returns.Credentials
}
