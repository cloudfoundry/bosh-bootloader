package fakes

import "github.com/cloudfoundry/bosh-bootloader/bosh"

type BOSHExecutor struct {
	CreateEnvCall struct {
		CallCount int
		Receives  struct {
			Input bosh.ExecutorInput
		}
		Returns struct {
			Output bosh.ExecutorOutput
			Error  error
		}
	}

	DeleteEnvCall struct {
		CallCount int
		Receives  struct {
			Input bosh.ExecutorInput
		}
		Returns struct {
			Output bosh.ExecutorOutput
			Error  error
		}
	}
}

func (e *BOSHExecutor) CreateEnv(input bosh.ExecutorInput) (bosh.ExecutorOutput, error) {
	e.CreateEnvCall.CallCount++
	e.CreateEnvCall.Receives.Input = input

	return e.CreateEnvCall.Returns.Output, e.CreateEnvCall.Returns.Error
}

func (e *BOSHExecutor) DeleteEnv(input bosh.ExecutorInput) (bosh.ExecutorOutput, error) {
	e.DeleteEnvCall.CallCount++
	e.DeleteEnvCall.Receives.Input = input

	return e.DeleteEnvCall.Returns.Output, e.DeleteEnvCall.Returns.Error
}
