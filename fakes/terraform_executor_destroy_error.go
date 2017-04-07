package fakes

type TerraformExecutorDestroyError struct {
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

func (t *TerraformExecutorDestroyError) Error() string {
	t.ErrorCall.CallCount++

	return t.ErrorCall.Returns
}

func (t *TerraformExecutorDestroyError) TFState() (string, error) {
	t.TFStateCall.CallCount++

	return t.TFStateCall.Returns.TFState, t.TFStateCall.Returns.Error
}
