package fakes

import "github.com/cloudfoundry/bosh-bootloader/terraform"

type OutputGenerator struct {
	GenerateCall struct {
		CallCount int
		Returns   struct {
			Outputs terraform.Outputs
			Error   error
		}
	}
}

func (o *OutputGenerator) Generate() (terraform.Outputs, error) {
	o.GenerateCall.CallCount++
	return o.GenerateCall.Returns.Outputs, o.GenerateCall.Returns.Error
}
