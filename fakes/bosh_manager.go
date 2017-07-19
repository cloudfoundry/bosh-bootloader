package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type BOSHManager struct {
	CreateJumpboxCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs map[string]interface{}
		}
		Returns struct {
			State storage.State
			Error error
		}
	}
	CreateDirectorCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs map[string]interface{}
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
			State            storage.State
			TerraformOutputs map[string]interface{}
		}
		Returns struct {
			Error error
		}
	}
	DeleteJumpboxCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs map[string]interface{}
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
	GetJumpboxDeploymentVarsCall struct {
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

func (b *BOSHManager) CreateJumpbox(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	b.CreateJumpboxCall.CallCount++
	b.CreateJumpboxCall.Receives.State = state
	b.GetDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	state.BOSH = b.CreateJumpboxCall.Returns.State.BOSH
	return state, b.CreateJumpboxCall.Returns.Error
}

func (b *BOSHManager) CreateDirector(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	b.CreateDirectorCall.CallCount++
	b.CreateDirectorCall.Receives.State = state
	b.GetDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	state.BOSH = b.CreateDirectorCall.Returns.State.BOSH
	return state, b.CreateDirectorCall.Returns.Error
}

func (b *BOSHManager) Delete(state storage.State, terraformOutputs map[string]interface{}) error {
	b.DeleteCall.CallCount++
	b.DeleteCall.Receives.State = state
	b.GetDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.DeleteCall.Returns.Error
}

func (b *BOSHManager) DeleteJumpbox(state storage.State, terraformOutputs map[string]interface{}) error {
	b.DeleteJumpboxCall.CallCount++
	b.DeleteJumpboxCall.Receives.State = state
	b.GetJumpboxDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.DeleteJumpboxCall.Returns.Error
}

func (b *BOSHManager) GetDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) (string, error) {
	b.GetDeploymentVarsCall.CallCount++
	b.GetDeploymentVarsCall.Receives.State = state
	b.GetDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.GetDeploymentVarsCall.Returns.Vars, b.GetDeploymentVarsCall.Returns.Error
}

func (b *BOSHManager) GetJumpboxDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) (string, error) {
	b.GetJumpboxDeploymentVarsCall.CallCount++
	b.GetJumpboxDeploymentVarsCall.Receives.State = state
	b.GetJumpboxDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.GetJumpboxDeploymentVarsCall.Returns.Vars, b.GetJumpboxDeploymentVarsCall.Returns.Error
}

func (b *BOSHManager) Version() (string, error) {
	b.VersionCall.CallCount++
	return b.VersionCall.Returns.Version, b.VersionCall.Returns.Error
}
