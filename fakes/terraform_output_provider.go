package fakes

import "github.com/cloudfoundry/bosh-bootloader/terraform"

type TerraformOutputProvider struct {
	GetCall struct {
		Returns struct {
			Outputs terraform.Outputs
			Error   error
		}

		Receives struct {
			tfState string
			lbType  string
		}
	}
}

func (t *TerraformOutputProvider) Get(tfState, lbType string) (terraform.Outputs, error) {
	t.GetCall.Receives.tfState = tfState
	t.GetCall.Receives.lbType = lbType
	return t.GetCall.Returns.Outputs, t.GetCall.Returns.Error
}
