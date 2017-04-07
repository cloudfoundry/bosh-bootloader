package fakes

type TerraformExecutorApplyError struct {
	ErrorCall struct {
		CallCount int
		Returns   string
	}

	TFStateCall struct {
		CallCount int
		Returns   struct {
			TFState string
			Error   error
		}
	}
}

func (t *TerraformExecutorApplyError) Error() string {
	t.ErrorCall.CallCount++

	return t.ErrorCall.Returns
}

func (t *TerraformExecutorApplyError) TFState() (string, error) {
	t.TFStateCall.CallCount++

	return t.TFStateCall.Returns.TFState, t.TFStateCall.Returns.Error
}
