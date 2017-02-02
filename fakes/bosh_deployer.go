package fakes

import "github.com/cloudfoundry/bosh-bootloader/bosh"

type BOSHDeployer struct {
	DeployCall struct {
		Receives struct {
			Input bosh.DeployInput
		}
		Returns struct {
			Output bosh.DeployOutput
			Error  error
		}
	}
}

func (d *BOSHDeployer) Deploy(input bosh.DeployInput) (bosh.DeployOutput, error) {
	d.DeployCall.Receives.Input = input

	return d.DeployCall.Returns.Output, d.DeployCall.Returns.Error
}
