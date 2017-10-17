package fakes

import "github.com/cloudfoundry/bosh-bootloader/bosh"

type BOSHExecutor struct {
	CreateEnvCall struct {
		CallCount int
		Receives  struct {
			Input bosh.CreateEnvInput
		}
		Returns struct {
			Error error
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
			Output bosh.InterpolateOutput
			Error  error
		}
	}

	DirectorInterpolateCall struct {
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

func (e *BOSHExecutor) CreateEnv(input bosh.CreateEnvInput) error {
	e.CreateEnvCall.CallCount++
	e.CreateEnvCall.Receives.Input = input

	return e.CreateEnvCall.Returns.Error
}

func (e *BOSHExecutor) DeleteEnv(input bosh.DeleteEnvInput) error {
	e.DeleteEnvCall.CallCount++
	e.DeleteEnvCall.Receives.Input = input

	return e.DeleteEnvCall.Returns.Error
}

func (e *BOSHExecutor) JumpboxInterpolate(input bosh.InterpolateInput) (bosh.InterpolateOutput, error) {
	e.JumpboxInterpolateCall.CallCount++
	e.JumpboxInterpolateCall.Receives.InterpolateInput = input

	return e.JumpboxInterpolateCall.Returns.Output, e.JumpboxInterpolateCall.Returns.Error
}

func (e *BOSHExecutor) DirectorInterpolate(input bosh.InterpolateInput) (bosh.InterpolateOutput, error) {
	e.DirectorInterpolateCall.CallCount++
	e.DirectorInterpolateCall.Receives.InterpolateInput = input

	return e.DirectorInterpolateCall.Returns.Output, e.DirectorInterpolateCall.Returns.Error
}

func (e *BOSHExecutor) Version() (string, error) {
	e.VersionCall.CallCount++
	return e.VersionCall.Returns.Version, e.VersionCall.Returns.Error
}
