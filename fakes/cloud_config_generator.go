package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

type CloudConfigGenerator struct {
	GenerateCall struct {
		Receives struct {
			CloudConfigInput bosh.CloudConfigInput
		}
		Returns struct {
			CloudConfig bosh.CloudConfig
			Error       error
		}
		CallCount int
	}
}

func (c *CloudConfigGenerator) Generate(cloudConfigInput bosh.CloudConfigInput) (bosh.CloudConfig, error) {
	c.GenerateCall.CallCount++
	c.GenerateCall.Receives.CloudConfigInput = cloudConfigInput

	return c.GenerateCall.Returns.CloudConfig, c.GenerateCall.Returns.Error
}
