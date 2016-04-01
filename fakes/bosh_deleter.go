package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type BOSHDeleter struct {
	DeleteCall struct {
		CallCount int
		Receives  struct {
			Input boshinit.DeployInput
		}
		Returns struct {
			Error error
		}
	}
}

func (d *BOSHDeleter) Delete(boshDeployInput boshinit.DeployInput) error {
	d.DeleteCall.CallCount++
	d.DeleteCall.Receives.Input = boshDeployInput

	return d.DeleteCall.Returns.Error
}
