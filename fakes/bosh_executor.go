package fakes

import "github.com/cloudfoundry/bosh-bootloader/bosh"

type BOSHExecutor struct {
	CreateEnvCall struct {
		CallCount int
		Receives  struct {
			Input bosh.CreateEnvInput
		}
		Returns struct {
			Output bosh.CreateEnvOutput
			Error  error
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

	JumpboxInterpolateCall struct {
		CallCount int
		Receives  struct {
			InterpolateInput bosh.InterpolateInput
		}
		Returns struct {
			Output bosh.JumpboxInterpolateOutput
			Error  error
		}
	}

	InterpolateCall struct {
		CallCount int
		Receives  struct {
			InterpolateInput bosh.InterpolateInput
		}
		Returns struct {
			Output bosh.InterpolateOutput
			Error  error
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

func (e *BOSHExecutor) CreateEnv(input bosh.CreateEnvInput) (bosh.CreateEnvOutput, error) {
	e.CreateEnvCall.CallCount++
	e.CreateEnvCall.Receives.Input = input

	return e.CreateEnvCall.Returns.Output, e.CreateEnvCall.Returns.Error
}

func (e *BOSHExecutor) DeleteEnv(input bosh.DeleteEnvInput) error {
	e.DeleteEnvCall.CallCount++
	e.DeleteEnvCall.Receives.Input = input

	return e.DeleteEnvCall.Returns.Error
}

func (e *BOSHExecutor) JumpboxInterpolate(input bosh.InterpolateInput) (bosh.JumpboxInterpolateOutput, error) {
	e.JumpboxInterpolateCall.CallCount++
	e.JumpboxInterpolateCall.Receives.InterpolateInput = input

	return e.JumpboxInterpolateCall.Returns.Output, e.JumpboxInterpolateCall.Returns.Error
}

func (e *BOSHExecutor) Interpolate(input bosh.InterpolateInput) (bosh.InterpolateOutput, error) {
	e.InterpolateCall.CallCount++
	e.InterpolateCall.Receives.InterpolateInput = input

	return e.InterpolateCall.Returns.Output, e.InterpolateCall.Returns.Error
}

func (e *BOSHExecutor) Version() (string, error) {
	e.VersionCall.CallCount++
	return e.VersionCall.Returns.Version, e.VersionCall.Returns.Error
}
