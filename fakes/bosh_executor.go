package fakes

import "github.com/cloudfoundry/bosh-bootloader/bosh"

type BOSHExecutor struct {
	CreateEnvCall struct {
		CallCount int
		Receives  struct {
			Input bosh.CreateEnvInput
		}
		Returns struct {
			Variables string
			Error     error
		}
	}

	DeleteEnvCall struct {
		CallCount int
		Receives  struct {
			Input bosh.DeleteEnvInput
		}
		Returns struct {
			Error error
		}
	}

	PlanJumpboxCall struct {
		CallCount int
		Receives  struct {
			InterpolateInput bosh.InterpolateInput
		}
		Returns struct {
			Error error
		}
	}

	PlanDirectorCall struct {
		CallCount int
		Receives  struct {
			InterpolateInput bosh.InterpolateInput
		}
		Returns struct {
			Error error
		}
	}

	WriteDeploymentVarsCall struct {
		CallCount int
		Receives  struct {
			Input bosh.CreateEnvInput
		}
		Returns struct {
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
}

func (e *BOSHExecutor) WriteDeploymentVars(input bosh.CreateEnvInput) error {
	e.WriteDeploymentVarsCall.CallCount++
	e.WriteDeploymentVarsCall.Receives.Input = input

	return e.WriteDeploymentVarsCall.Returns.Error
}

func (e *BOSHExecutor) CreateEnv(input bosh.CreateEnvInput) (string, error) {
	e.CreateEnvCall.CallCount++
	e.CreateEnvCall.Receives.Input = input

	return e.CreateEnvCall.Returns.Variables, e.CreateEnvCall.Returns.Error
}

func (e *BOSHExecutor) DeleteEnv(input bosh.DeleteEnvInput) error {
	e.DeleteEnvCall.CallCount++
	e.DeleteEnvCall.Receives.Input = input

	return e.DeleteEnvCall.Returns.Error
}

func (e *BOSHExecutor) PlanJumpbox(input bosh.InterpolateInput) error {
	e.PlanJumpboxCall.CallCount++
	e.PlanJumpboxCall.Receives.InterpolateInput = input

	return e.PlanJumpboxCall.Returns.Error
}

func (e *BOSHExecutor) PlanDirector(input bosh.InterpolateInput) error {
	e.PlanDirectorCall.CallCount++
	e.PlanDirectorCall.Receives.InterpolateInput = input

	return e.PlanDirectorCall.Returns.Error
}

func (e *BOSHExecutor) Version() (string, error) {
	e.VersionCall.CallCount++
	return e.VersionCall.Returns.Version, e.VersionCall.Returns.Error
}
