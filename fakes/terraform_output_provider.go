package fakes

import "github.com/cloudfoundry/bosh-bootloader/terraform"

type TerraformOutputProvider struct {
	GetCall struct {
		CallCount int
		Returns   struct {
			Outputs terraform.Outputs
			Error   error
		}

		Receives struct {
			TFState      string
			LBType       string
			DomainExists bool
		}
	}
}

func (t *TerraformOutputProvider) Get(tfState, lbType string, domainExists bool) (terraform.Outputs, error) {
	t.GetCall.CallCount++
	t.GetCall.Receives.TFState = tfState
	t.GetCall.Receives.LBType = lbType
	t.GetCall.Receives.DomainExists = domainExists
	return t.GetCall.Returns.Outputs, t.GetCall.Returns.Error
}
