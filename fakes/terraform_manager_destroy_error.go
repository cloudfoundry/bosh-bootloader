package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type TerraformManagerDestroyError struct {
	BBLStateCall struct {
		CallCount int
		Returns   struct {
			BBLState storage.State
			Error    error
		}
	}
	ErrorCall struct {
		CallCount int
		Returns   string
	}
}

func (t *TerraformManagerDestroyError) BBLState() (storage.State, error) {
	t.BBLStateCall.CallCount++
	return t.BBLStateCall.Returns.BBLState, t.BBLStateCall.Returns.Error
}

func (t *TerraformManagerDestroyError) Error() string {
	t.ErrorCall.CallCount++
	return t.ErrorCall.Returns
}
