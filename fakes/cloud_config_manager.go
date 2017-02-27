package fakes

import "github.com/cloudfoundry/bosh-bootloader/bosh"

type CloudConfigManager struct {
	UpdateCall struct {
		CallCount int
		Receives  struct {
			CloudConfigInput bosh.CloudConfigInput
			BOSHClient       bosh.Client
		}
		Returns struct {
			Error error
		}
	}
}

func (c *CloudConfigManager) Update(cloudConfigInput bosh.CloudConfigInput, boshClient bosh.Client) error {
	c.UpdateCall.CallCount++
	c.UpdateCall.Receives.CloudConfigInput = cloudConfigInput
	c.UpdateCall.Receives.BOSHClient = boshClient
	return c.UpdateCall.Returns.Error
}
