package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type BOSHExecutor struct {
	CreateEnvCall struct {
		CallCount int
		Receives  struct {
			DirInput bosh.DirInput
			State    storage.State
		}
		Returns struct {
			Variables string
			Error     error
		}
	}

	DeleteEnvCall struct {
		CallCount int
		Receives  struct {
			DirInput bosh.DirInput
			State    storage.State
		}
		Returns struct {
			Error error
		}
	}

	PlanJumpboxCall struct {
		CallCount int
		Receives  struct {
			DirInput      bosh.DirInput
			DeploymentDir string
			Iaas          string
		}
		Returns struct {
			Error error
		}
	}

	PlanDirectorCall struct {
		CallCount int
		Receives  struct {
			DirInput      bosh.DirInput
			DeploymentDir string
			Iaas          string
		}
		Returns struct {
			Error error
		}
	}

	WriteDeploymentVarsCall struct {
		CallCount int
		Receives  struct {
			DirInput       bosh.DirInput
			DeploymentVars string
		}
		Returns struct {
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
}

func (e *BOSHExecutor) WriteDeploymentVars(input bosh.DirInput, deploymentVars string) error {
	e.WriteDeploymentVarsCall.CallCount++
	e.WriteDeploymentVarsCall.Receives.DirInput = input
	e.WriteDeploymentVarsCall.Receives.DeploymentVars = deploymentVars

	return e.WriteDeploymentVarsCall.Returns.Error
}

func (e *BOSHExecutor) CreateEnv(input bosh.DirInput, state storage.State) (string, error) {
	e.CreateEnvCall.CallCount++
	e.CreateEnvCall.Receives.DirInput = input
	e.CreateEnvCall.Receives.State = state

	return e.CreateEnvCall.Returns.Variables, e.CreateEnvCall.Returns.Error
}

func (e *BOSHExecutor) DeleteEnv(input bosh.DirInput, state storage.State) error {
	e.DeleteEnvCall.CallCount++
	e.DeleteEnvCall.Receives.DirInput = input
	e.DeleteEnvCall.Receives.State = state

	return e.DeleteEnvCall.Returns.Error
}

func (e *BOSHExecutor) PlanJumpbox(input bosh.DirInput, deploymentDir, iaas string) error {
	e.PlanJumpboxCall.CallCount++
	e.PlanJumpboxCall.Receives.DirInput = input
	e.PlanJumpboxCall.Receives.DeploymentDir = deploymentDir
	e.PlanJumpboxCall.Receives.Iaas = iaas

	return e.PlanJumpboxCall.Returns.Error
}

func (e *BOSHExecutor) PlanDirector(input bosh.DirInput, deploymentDir, iaas string) error {
	e.PlanDirectorCall.CallCount++
	e.PlanDirectorCall.Receives.DirInput = input
	e.PlanDirectorCall.Receives.DeploymentDir = deploymentDir
	e.PlanDirectorCall.Receives.Iaas = iaas

	return e.PlanDirectorCall.Returns.Error
}

func (e *BOSHExecutor) Path() string {
	e.PathCall.CallCount++
	return e.PathCall.Returns.Path
}

func (e *BOSHExecutor) Version() (string, error) {
	e.VersionCall.CallCount++
	return e.VersionCall.Returns.Version, e.VersionCall.Returns.Error
}
