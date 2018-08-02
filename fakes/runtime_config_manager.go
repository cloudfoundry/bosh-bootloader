package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type RuntimeConfigManager struct {
	UpdateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
}

func (c *RuntimeConfigManager) Update(state storage.State) error {
	c.UpdateCall.CallCount++
	c.UpdateCall.Receives.State = state
	return c.UpdateCall.Returns.Error
}
