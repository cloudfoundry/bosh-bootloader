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
			TFState string
			LBType  string
		}
	}
}

func (t *TerraformOutputProvider) Get(tfState, lbType string) (terraform.Outputs, error) {
	t.GetCall.CallCount++
	t.GetCall.Receives.TFState = tfState
	t.GetCall.Receives.LBType = lbType
	return t.GetCall.Returns.Outputs, t.GetCall.Returns.Error
}
