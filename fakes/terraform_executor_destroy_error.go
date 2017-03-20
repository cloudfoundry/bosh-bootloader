package fakes

type TerraformExecutorDestroyError struct {
	ErrorCall struct {
		CallCount int
		Returns   string
	}
}

func (t *TerraformExecutorDestroyError) Error() string {
	t.ErrorCall.CallCount++

	return t.ErrorCall.Returns
}
