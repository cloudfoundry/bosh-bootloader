package fakes

import "github.com/cloudfoundry/bosh-bootloader/boshinit"

type BOSHInitDeployer struct {
	DeployCall struct {
		Receives struct {
			Input boshinit.DeployInput
		}
		Returns struct {
			Output boshinit.DeployOutput
			Error  error
		}
	}
}

func (d *BOSHInitDeployer) Deploy(input boshinit.DeployInput) (boshinit.DeployOutput, error) {
	d.DeployCall.Receives.Input = input

	return d.DeployCall.Returns.Output, d.DeployCall.Returns.Error
}
