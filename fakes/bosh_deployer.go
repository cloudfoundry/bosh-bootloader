package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"

type BOSHDeployer struct {
	DeployCall struct {
		Receives struct {
			Input unsupported.BOSHDeployInput
		}
		Returns struct {
			Output unsupported.BOSHDeployOutput
			Error  error
		}
	}
}

func (d *BOSHDeployer) Deploy(input unsupported.BOSHDeployInput) (unsupported.BOSHDeployOutput, error) {
	d.DeployCall.Receives.Input = input

	return d.DeployCall.Returns.Output, d.DeployCall.Returns.Error
}
