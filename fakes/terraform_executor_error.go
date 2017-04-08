package fakes

type TerraformExecutorError struct {
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

func (t *TerraformExecutorError) Error() string {
	t.ErrorCall.CallCount++

	return t.ErrorCall.Returns
}

func (t *TerraformExecutorError) TFState() (string, error) {
	t.TFStateCall.CallCount++

	return t.TFStateCall.Returns.TFState, t.TFStateCall.Returns.Error
}
