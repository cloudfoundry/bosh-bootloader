package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type TerraformManagerApplyError struct {
	BBLStateCall struct {
		CallCount int
		Returns   storage.State
	}
	ErrorCall struct {
		CallCount int
		Returns   string
	}
}

func (t *TerraformManagerApplyError) BBLState() storage.State {
	t.BBLStateCall.CallCount++
	return t.BBLStateCall.Returns
}

func (t *TerraformManagerApplyError) Error() string {
	t.ErrorCall.CallCount++
	return t.ErrorCall.Returns
}
