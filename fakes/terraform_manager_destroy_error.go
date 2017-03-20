package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type TerraformManagerDestroyError struct {
	BBLStateCall struct {
		CallCount int
		Returns   storage.State
	}
	ErrorCall struct {
		CallCount int
		Returns   string
	}
}

func (t *TerraformManagerDestroyError) BBLState() storage.State {
	t.BBLStateCall.CallCount++
	return t.BBLStateCall.Returns
}

func (t *TerraformManagerDestroyError) Error() string {
	t.ErrorCall.CallCount++
	return t.ErrorCall.Returns
}
