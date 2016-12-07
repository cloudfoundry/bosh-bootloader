package fakes

type TerraformOutputer struct {
	GetCall struct {
		Stub      func(string) (string, error)
		CallCount int
		Receives  struct {
			TFState    string
			OutputName string
		}
		Returns struct {
			Output string
			Error  error
		}
	}
}

func (t *TerraformOutputer) Get(tfState, outputName string) (string, error) {
	t.GetCall.CallCount++
	t.GetCall.Receives.TFState = tfState
	t.GetCall.Receives.OutputName = outputName

	if t.GetCall.Stub != nil {
		return t.GetCall.Stub(outputName)
	}

	return t.GetCall.Returns.Output, t.GetCall.Returns.Error
}
