package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

type BoshCloudConfigurator struct {
	ConfigureCall struct {
		CallCount int
		Receives  struct {
			Stack cloudformation.Stack
			AZs   []string
		}
		Returns struct {
			Error error
		}
	}
}

func (b *BoshCloudConfigurator) Configure(stack cloudformation.Stack, azs []string) error {
	b.ConfigureCall.CallCount++
	b.ConfigureCall.Receives.Stack = stack
	b.ConfigureCall.Receives.AZs = azs

	return b.ConfigureCall.Returns.Error
}
