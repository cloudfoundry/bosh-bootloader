package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
)

type BoshCloudConfigurator struct {
	ConfigureCall struct {
		CallCount int
		Receives  struct {
			Stack  cloudformation.Stack
			AZs    []string
			Client bosh.Client
		}
		Returns struct {
			Error error
		}
	}
}

func (b *BoshCloudConfigurator) Configure(stack cloudformation.Stack, azs []string, client bosh.Client) error {
	b.ConfigureCall.CallCount++
	b.ConfigureCall.Receives.Stack = stack
	b.ConfigureCall.Receives.AZs = azs
	b.ConfigureCall.Receives.Client = client

	return b.ConfigureCall.Returns.Error
}
