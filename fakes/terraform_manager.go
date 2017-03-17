package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type TerraformManager struct {
	DestroyCall struct {
		CallCount int
		Receives  struct {
			BBLState storage.State
		}
		Returns struct {
			BBLState storage.State
			Error    error
		}
	}
	GetOutputsCall struct {
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

func (t *TerraformManager) Destroy(bblState storage.State) (storage.State, error) {
	t.DestroyCall.CallCount++
	t.DestroyCall.Receives.BBLState = bblState

	return t.DestroyCall.Returns.BBLState, t.DestroyCall.Returns.Error
}

func (t *TerraformManager) GetOutputs(tfState, lbType string, domainExists bool) (terraform.Outputs, error) {
	t.GetOutputsCall.CallCount++
	t.GetOutputsCall.Receives.TFState = tfState
	t.GetOutputsCall.Receives.LBType = lbType
	t.GetOutputsCall.Receives.DomainExists = domainExists

	return t.GetOutputsCall.Returns.Outputs, t.GetOutputsCall.Returns.Error
}
