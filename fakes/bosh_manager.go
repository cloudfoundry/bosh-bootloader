package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type BOSHManager struct {
	CreateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
	VersionCall struct {
		CallCount int
		Returns   struct {
			Version string
			Error   error
		}
	}
	DeleteCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
	GetDeploymentVarsCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs map[string]interface{}
		}
		Returns struct {
			Vars  string
			Error error
		}
	}
}

func (b *BOSHManager) Create(state storage.State) (storage.State, error) {
	b.CreateCall.CallCount++
	b.CreateCall.Receives.State = state
	state.BOSH = b.CreateCall.Returns.State.BOSH
	return state, b.CreateCall.Returns.Error
}

func (b *BOSHManager) Delete(state storage.State) error {
	b.DeleteCall.CallCount++
	b.DeleteCall.Receives.State = state
	return b.DeleteCall.Returns.Error
}

func (b *BOSHManager) GetDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) (string, error) {
	b.GetDeploymentVarsCall.CallCount++
	b.GetDeploymentVarsCall.Receives.State = state
	b.GetDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.GetDeploymentVarsCall.Returns.Vars, b.GetDeploymentVarsCall.Returns.Error
}

func (b *BOSHManager) Version() (string, error) {
	b.VersionCall.CallCount++
	return b.VersionCall.Returns.Version, b.VersionCall.Returns.Error
}
