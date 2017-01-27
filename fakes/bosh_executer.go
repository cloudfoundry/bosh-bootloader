package fakes

import "github.com/cloudfoundry/bosh-bootloader/bosh"

type BOSHExecutor struct {
	ExecuteCall struct {
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

func (e *BOSHExecutor) Execute(input bosh.ExecutorInput) (bosh.ExecutorOutput, error) {
	e.ExecuteCall.CallCount++
	e.ExecuteCall.Receives.Input = input

	return e.ExecuteCall.Returns.Output, e.ExecuteCall.Returns.Error
}
