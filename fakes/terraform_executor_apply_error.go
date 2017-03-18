package fakes

type TerraformExecutorApplyError struct {
	ErrorCall struct {
		CallCount int
		Returns   string
	}
}

func (t *TerraformExecutorApplyError) Error() string {
	t.ErrorCall.CallCount++

	return t.ErrorCall.Returns
}
