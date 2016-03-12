package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

type BOSHDeployer struct {
	DeployCall struct {
		Receives struct {
			Input unsupported.BOSHDeployInput
		}
		Returns struct {
			BOSHInitState      boshinit.State
			DirectorSSLKeyPair ssl.KeyPair
			Error              error
		}
	}
}

func (d *BOSHDeployer) Deploy(input unsupported.BOSHDeployInput) (boshinit.State, ssl.KeyPair, error) {
	d.DeployCall.Receives.Input = input

	return d.DeployCall.Returns.BOSHInitState, d.DeployCall.Returns.DirectorSSLKeyPair, d.DeployCall.Returns.Error
}
