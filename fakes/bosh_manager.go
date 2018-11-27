package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type BOSHManager struct {
	InitializeJumpboxCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
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
	InitializeDirectorCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
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
	PathCall struct {
		CallCount int
		Returns   struct {
			Path string
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
	CleanUpDirectorCall struct {
		CallCount int
		Receives  struct {
			State storage.State
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

func (b *BOSHManager) InitializeJumpbox(state storage.State) error {
	b.InitializeJumpboxCall.CallCount++
	b.InitializeJumpboxCall.Receives.State = state
	return b.InitializeJumpboxCall.Returns.Error
}

func (b *BOSHManager) CreateJumpbox(state storage.State, terraformOutputs terraform.Outputs) (storage.State, error) {
	b.CreateJumpboxCall.CallCount++
	b.CreateJumpboxCall.Receives.State = state
	b.CreateJumpboxCall.Receives.TerraformOutputs = terraformOutputs
	return b.CreateJumpboxCall.Returns.State, b.CreateJumpboxCall.Returns.Error
}

func (b *BOSHManager) InitializeDirector(state storage.State) error {
	b.InitializeDirectorCall.CallCount++
	b.InitializeDirectorCall.Receives.State = state
	return b.InitializeDirectorCall.Returns.Error
}

func (b *BOSHManager) CreateDirector(state storage.State, terraformOutputs terraform.Outputs) (storage.State, error) {
	b.CreateDirectorCall.CallCount++
	b.CreateDirectorCall.Receives.State = state
	b.CreateDirectorCall.Receives.TerraformOutputs = terraformOutputs
	return b.CreateDirectorCall.Returns.State, b.CreateDirectorCall.Returns.Error
}

func (b *BOSHManager) DeleteDirector(state storage.State, terraformOutputs terraform.Outputs) error {
	b.DeleteDirectorCall.CallCount++
	b.DeleteDirectorCall.Receives.State = state
	b.DeleteDirectorCall.Receives.TerraformOutputs = terraformOutputs
	return b.DeleteDirectorCall.Returns.Error
}

func (b *BOSHManager) CleanUpDirector(state storage.State) error {
	b.CleanUpDirectorCall.CallCount++
	b.CleanUpDirectorCall.Receives.State = state
	return b.CleanUpDirectorCall.Returns.Error
}

func (b *BOSHManager) DeleteJumpbox(state storage.State, terraformOutputs terraform.Outputs) error {
	b.DeleteJumpboxCall.CallCount++
	b.DeleteJumpboxCall.Receives.State = state
	b.DeleteJumpboxCall.Receives.TerraformOutputs = terraformOutputs
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

func (b *BOSHManager) Path() string {
	b.PathCall.CallCount++
	return b.PathCall.Returns.Path
}

func (b *BOSHManager) Version() (string, error) {
	b.VersionCall.CallCount++
	return b.VersionCall.Returns.Version, b.VersionCall.Returns.Error
}
