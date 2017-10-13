package fakes

import "github.com/cloudfoundry/bosh-bootloader/terraform"

type OutputGenerator struct {
	GenerateCall struct {
		CallCount int
		Receives  struct {
			TFState string
		}
		Returns struct {
			Outputs terraform.Outputs
			Error   error
		}
	}
}

func (o *OutputGenerator) Generate(tfState string) (terraform.Outputs, error) {
	o.GenerateCall.CallCount++
	o.GenerateCall.Receives.TFState = tfState
	return o.GenerateCall.Returns.Outputs, o.GenerateCall.Returns.Error
}
