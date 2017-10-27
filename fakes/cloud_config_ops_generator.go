package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type CloudConfigOpsGenerator struct {
	GenerateCall struct {
		Receives struct {
			State storage.State
		}
		Returns struct {
			OpsYAML string
			Error   error
		}
	}
	GenerateVarsCall struct {
		Receives struct {
			State storage.State
		}
		Returns struct {
			VarsYAML string
			Error    error
		}
	}
}

func (c *CloudConfigOpsGenerator) Generate(state storage.State) (string, error) {
	c.GenerateCall.Receives.State = state
	return c.GenerateCall.Returns.OpsYAML, c.GenerateCall.Returns.Error
}

func (c *CloudConfigOpsGenerator) GenerateVars(state storage.State) (string, error) {
	c.GenerateVarsCall.Receives.State = state
	return c.GenerateVarsCall.Returns.VarsYAML, c.GenerateVarsCall.Returns.Error
}
