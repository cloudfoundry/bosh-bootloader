package fakes

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
)

type BoshCloudConfigurator struct {
	ConfigureCall struct {
		CallCount int
		Receives  struct {
			Stack cloudformation.Stack
			AZs   []string
		}
		Returns struct {
			CloudConfigInput bosh.CloudConfigInput
		}
	}
}

func (b *BoshCloudConfigurator) Configure(stack cloudformation.Stack, azs []string) bosh.CloudConfigInput {
	b.ConfigureCall.CallCount++
	b.ConfigureCall.Receives.Stack = stack
	b.ConfigureCall.Receives.AZs = azs

	return b.ConfigureCall.Returns.CloudConfigInput
}
