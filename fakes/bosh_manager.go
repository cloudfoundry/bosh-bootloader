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
	GetDirectorDeploymentVarsCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs map[string]interface{}
		}
		Returns struct {
			Vars string
		}
	}
	GetJumpboxDeploymentVarsCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs map[string]interface{}
		}
		Returns struct {
			Vars string
		}
	}
}

func (b *BOSHManager) CreateJumpbox(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	b.CreateJumpboxCall.CallCount++
	b.CreateJumpboxCall.Receives.State = state
	b.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.CreateJumpboxCall.Returns.State, b.CreateJumpboxCall.Returns.Error
}

func (b *BOSHManager) CreateDirector(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	b.CreateDirectorCall.CallCount++
	b.CreateDirectorCall.Receives.State = state
	b.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.CreateDirectorCall.Returns.State, b.CreateDirectorCall.Returns.Error
}

func (b *BOSHManager) Delete(state storage.State, terraformOutputs map[string]interface{}) error {
	b.DeleteCall.CallCount++
	b.DeleteCall.Receives.State = state
	b.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.DeleteCall.Returns.Error
}

func (b *BOSHManager) DeleteJumpbox(state storage.State, terraformOutputs map[string]interface{}) error {
	b.DeleteJumpboxCall.CallCount++
	b.DeleteJumpboxCall.Receives.State = state
	b.GetJumpboxDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.DeleteJumpboxCall.Returns.Error
}

func (b *BOSHManager) GetDirectorDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) string {
	b.GetDirectorDeploymentVarsCall.CallCount++
	b.GetDirectorDeploymentVarsCall.Receives.State = state
	b.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.GetDirectorDeploymentVarsCall.Returns.Vars
}

func (b *BOSHManager) GetJumpboxDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) string {
	b.GetJumpboxDeploymentVarsCall.CallCount++
	b.GetJumpboxDeploymentVarsCall.Receives.State = state
	b.GetJumpboxDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.GetJumpboxDeploymentVarsCall.Returns.Vars
}

func (b *BOSHManager) Version() (string, error) {
	b.VersionCall.CallCount++
	return b.VersionCall.Returns.Version, b.VersionCall.Returns.Error
}
