package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type LBArgsHandler struct {
	GetLBStateCall struct {
		CallCount int
		Returns   struct {
			LB    storage.LB
			Error error
		}
		Receives struct {
			IAAS   string
			Config commands.CreateLBsConfig
		}
	}
	MergeCall struct {
		CallCount int
		Returns   struct {
			LB storage.LB
		}
		Receives struct {
			New storage.LB
			Old storage.LB
		}
	}
}

func (c *LBArgsHandler) GetLBState(iaas string, config commands.CreateLBsConfig) (storage.LB, error) {
	c.GetLBStateCall.CallCount++
	c.GetLBStateCall.Receives.Config = config
	c.GetLBStateCall.Receives.IAAS = iaas
	return c.GetLBStateCall.Returns.LB, c.GetLBStateCall.Returns.Error
}

func (c *LBArgsHandler) Merge(new storage.LB, old storage.LB) storage.LB {
	c.MergeCall.CallCount++
	c.MergeCall.Receives.New = new
	c.MergeCall.Receives.Old = old
	return c.MergeCall.Returns.LB
}
