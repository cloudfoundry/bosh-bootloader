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
	InitializeCall struct {
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

func (c *RuntimeConfigManager) Initialize(state storage.State) error {
	c.InitializeCall.CallCount++
	c.InitializeCall.Receives.State = state
	return c.InitializeCall.Returns.Error
}
