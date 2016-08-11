package fakes

type EnvIDGenerator struct {
	GenerateCall struct {
		CallCount int
		Returns   struct {
			EnvID string
			Error error
		}
	}
}

func (e *EnvIDGenerator) Generate() (string, error) {
	e.GenerateCall.CallCount++
	return e.GenerateCall.Returns.EnvID, e.GenerateCall.Returns.Error
}
