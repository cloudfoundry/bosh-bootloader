package fakes

type StringGenerator struct {
	GenerateCall struct {
		Receives struct {
			Prefix string
			Length int
		}
		Returns struct {
			String string
			Error  error
		}
		Stub      func(string, int) (string, error)
		CallCount int
	}
}

func (s *StringGenerator) Generate(prefix string, length int) (string, error) {
	defer func() { s.GenerateCall.CallCount++ }()
	s.GenerateCall.Receives.Length = length
	s.GenerateCall.Receives.Prefix = prefix

	if s.GenerateCall.Stub != nil {
		return s.GenerateCall.Stub(prefix, length)
	}

	return s.GenerateCall.Returns.String, s.GenerateCall.Returns.Error
}
