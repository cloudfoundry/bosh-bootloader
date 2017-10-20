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

	JumpboxCreateEnvArgsCall struct {
		CallCount int
		Receives  struct {
			InterpolateInput bosh.InterpolateInput
		}
		Returns struct {
			Error error
		}
	}

	DirectorCreateEnvArgsCall struct {
		CallCount int
		Receives  struct {
			InterpolateInput bosh.InterpolateInput
		}
		Returns struct {
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

func (e *BOSHExecutor) JumpboxCreateEnvArgs(input bosh.InterpolateInput) error {
	e.JumpboxCreateEnvArgsCall.CallCount++
	e.JumpboxCreateEnvArgsCall.Receives.InterpolateInput = input

	return e.JumpboxCreateEnvArgsCall.Returns.Error
}

func (e *BOSHExecutor) DirectorCreateEnvArgs(input bosh.InterpolateInput) error {
	e.DirectorCreateEnvArgsCall.CallCount++
	e.DirectorCreateEnvArgsCall.Receives.InterpolateInput = input

	return e.DirectorCreateEnvArgsCall.Returns.Error
}

func (e *BOSHExecutor) Version() (string, error) {
	e.VersionCall.CallCount++
	return e.VersionCall.Returns.Version, e.VersionCall.Returns.Error
}
