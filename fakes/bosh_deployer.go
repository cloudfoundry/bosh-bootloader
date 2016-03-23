package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type BOSHDeployer struct {
	DeployCall struct {
		Receives struct {
			Input boshinit.BOSHDeployInput
		}
		Returns struct {
			Output boshinit.BOSHDeployOutput
			Error  error
		}
	}
}

func (d *BOSHDeployer) Deploy(input boshinit.BOSHDeployInput) (boshinit.BOSHDeployOutput, error) {
	d.DeployCall.Receives.Input = input

	return d.DeployCall.Returns.Output, d.DeployCall.Returns.Error
}
