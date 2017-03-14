package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type CloudConfigManager struct {
	UpdateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
	GenerateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			CloudConfig string
			Error       error
		}
	}
}

func (c *CloudConfigManager) Update(state storage.State) error {
	c.UpdateCall.CallCount++
	c.UpdateCall.Receives.State = state
	return c.UpdateCall.Returns.Error
}

func (c *CloudConfigManager) Generate(state storage.State) (string, error) {
	c.GenerateCall.CallCount++
	c.GenerateCall.Receives.State = state
	return c.GenerateCall.Returns.CloudConfig, c.GenerateCall.Returns.Error
}
