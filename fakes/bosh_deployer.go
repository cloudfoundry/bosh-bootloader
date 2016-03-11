package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

type BOSHDeployer struct {
	DeployCall struct {
		Receives struct {
			Stack                cloudformation.Stack
			CloudformationClient cloudformation.Client
			AWSRegion            string
			KeyPairName          string
			DirectorSSLKeyPair   ssl.KeyPair
		}
		Returns struct {
			DirectorSSLKeyPair ssl.KeyPair
			Error              error
		}
	}
}

func (d *BOSHDeployer) Deploy(stack cloudformation.Stack, client cloudformation.Client, region, keyPairName string, directorSSLKeyPair ssl.KeyPair) (ssl.KeyPair, error) {
	d.DeployCall.Receives.Stack = stack
	d.DeployCall.Receives.CloudformationClient = client
	d.DeployCall.Receives.AWSRegion = region
	d.DeployCall.Receives.KeyPairName = keyPairName
	d.DeployCall.Receives.DirectorSSLKeyPair = directorSSLKeyPair

	return d.DeployCall.Returns.DirectorSSLKeyPair, d.DeployCall.Returns.Error
}
