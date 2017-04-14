package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TerraformManager struct {
	DestroyCall struct {
		CallCount int
		Receives  struct {
			BBLState storage.State
		}
		Returns struct {
			BBLState storage.State
			Error    error
		}
	}
	GetOutputsCall struct {
		CallCount int
		Receives  struct {
			BBLState storage.State
		}
		Returns struct {
			Outputs map[string]interface{}
			Error   error
		}
	}
	VersionCall struct {
		CallCount int
		Returns   struct {
			Version string
			Error   error
		}
	}
	ApplyCall struct {
		CallCount int
		Receives  struct {
			BBLState storage.State
		}
		Returns struct {
			BBLState storage.State
			Error    error
		}
	}
	ValidateVersionCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (t *TerraformManager) Apply(bblState storage.State) (storage.State, error) {
	t.ApplyCall.CallCount++
	t.ApplyCall.Receives.BBLState = bblState

	return t.ApplyCall.Returns.BBLState, t.ApplyCall.Returns.Error
}

func (t *TerraformManager) Destroy(bblState storage.State) (storage.State, error) {
	t.DestroyCall.CallCount++
	t.DestroyCall.Receives.BBLState = bblState

	return t.DestroyCall.Returns.BBLState, t.DestroyCall.Returns.Error
}

func (t *TerraformManager) GetOutputs(bblState storage.State) (map[string]interface{}, error) {
	t.GetOutputsCall.CallCount++
	t.GetOutputsCall.Receives.BBLState = bblState

	return t.GetOutputsCall.Returns.Outputs, t.GetOutputsCall.Returns.Error
}

func (t *TerraformManager) Version() (string, error) {
	t.VersionCall.CallCount++
	return t.VersionCall.Returns.Version, t.VersionCall.Returns.Error
}

func (t *TerraformManager) ValidateVersion() error {
	t.ValidateVersionCall.CallCount++
	return t.ValidateVersionCall.Returns.Error
}
