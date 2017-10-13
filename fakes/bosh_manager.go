package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type BOSHManager struct {
	CreateJumpboxCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs terraform.Outputs
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
			TerraformOutputs terraform.Outputs
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
	DeleteDirectorCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs terraform.Outputs
		}
		Returns struct {
			Error error
		}
	}
	DeleteJumpboxCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs terraform.Outputs
		}
		Returns struct {
			Error error
		}
	}
	GetDirectorDeploymentVarsCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs terraform.Outputs
		}
		Returns struct {
			Vars string
		}
	}
	GetJumpboxDeploymentVarsCall struct {
		CallCount int
		Receives  struct {
			State            storage.State
			TerraformOutputs terraform.Outputs
		}
		Returns struct {
			Vars string
		}
	}
}

func (b *BOSHManager) CreateJumpbox(state storage.State, terraformOutputs terraform.Outputs) (storage.State, error) {
	b.CreateJumpboxCall.CallCount++
	b.CreateJumpboxCall.Receives.State = state
	b.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.CreateJumpboxCall.Returns.State, b.CreateJumpboxCall.Returns.Error
}

func (b *BOSHManager) CreateDirector(state storage.State, terraformOutputs terraform.Outputs) (storage.State, error) {
	b.CreateDirectorCall.CallCount++
	b.CreateDirectorCall.Receives.State = state
	b.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.CreateDirectorCall.Returns.State, b.CreateDirectorCall.Returns.Error
}

func (b *BOSHManager) DeleteDirector(state storage.State, terraformOutputs terraform.Outputs) error {
	b.DeleteDirectorCall.CallCount++
	b.DeleteDirectorCall.Receives.State = state
	b.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.DeleteDirectorCall.Returns.Error
}

func (b *BOSHManager) DeleteJumpbox(state storage.State, terraformOutputs terraform.Outputs) error {
	b.DeleteJumpboxCall.CallCount++
	b.DeleteJumpboxCall.Receives.State = state
	b.GetJumpboxDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.DeleteJumpboxCall.Returns.Error
}

func (b *BOSHManager) GetDirectorDeploymentVars(state storage.State, terraformOutputs terraform.Outputs) string {
	b.GetDirectorDeploymentVarsCall.CallCount++
	b.GetDirectorDeploymentVarsCall.Receives.State = state
	b.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.GetDirectorDeploymentVarsCall.Returns.Vars
}

func (b *BOSHManager) GetJumpboxDeploymentVars(state storage.State, terraformOutputs terraform.Outputs) string {
	b.GetJumpboxDeploymentVarsCall.CallCount++
	b.GetJumpboxDeploymentVarsCall.Receives.State = state
	b.GetJumpboxDeploymentVarsCall.Receives.TerraformOutputs = terraformOutputs
	return b.GetJumpboxDeploymentVarsCall.Returns.Vars
}

func (b *BOSHManager) Version() (string, error) {
	b.VersionCall.CallCount++
	return b.VersionCall.Returns.Version, b.VersionCall.Returns.Error
}
