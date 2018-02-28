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
	InitializeCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
	GenerateVarsCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Error error
		}
	}
	InterpolateCall struct {
		CallCount int
		Returns   struct {
			CloudConfig string
			Error       error
		}
	}
	IsPresentCloudConfigCall struct {
		CallCount int
		Returns   struct {
			IsPresent bool
		}
	}
	IsPresentCloudConfigVarsCall struct {
		CallCount int
		Returns   struct {
			IsPresent bool
		}
	}
}

func (c *CloudConfigManager) Update(state storage.State) error {
	c.UpdateCall.CallCount++
	c.UpdateCall.Receives.State = state
	return c.UpdateCall.Returns.Error
}

func (c *CloudConfigManager) Initialize(state storage.State) error {
	c.InitializeCall.CallCount++
	c.InitializeCall.Receives.State = state
	return c.InitializeCall.Returns.Error
}

func (c *CloudConfigManager) GenerateVars(state storage.State) error {
	c.GenerateVarsCall.CallCount++
	c.GenerateVarsCall.Receives.State = state
	return c.GenerateVarsCall.Returns.Error
}

func (c *CloudConfigManager) Interpolate() (string, error) {
	c.InterpolateCall.CallCount++
	return c.InterpolateCall.Returns.CloudConfig, c.InterpolateCall.Returns.Error
}

func (c *CloudConfigManager) IsPresentCloudConfig() bool {
	c.IsPresentCloudConfigCall.CallCount++
	return c.IsPresentCloudConfigCall.Returns.IsPresent
}

func (c *CloudConfigManager) IsPresentCloudConfigVars() bool {
	c.IsPresentCloudConfigVarsCall.CallCount++
	return c.IsPresentCloudConfigVarsCall.Returns.IsPresent
}
