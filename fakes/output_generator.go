package fakes

type OutputGenerator struct {
	GenerateCall struct {
		CallCount int
		Receives  struct {
			TFState string
		}
		Returns struct {
			Outputs map[string]interface{}
			Error   error
		}
	}
}

func (o *OutputGenerator) Generate(tfState string) (map[string]interface{}, error) {
	o.GenerateCall.CallCount++
	o.GenerateCall.Receives.TFState = tfState
	return o.GenerateCall.Returns.Outputs, o.GenerateCall.Returns.Error
}
