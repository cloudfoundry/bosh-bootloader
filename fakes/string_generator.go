package fakes

type StringGenerator struct {
	GenerateCall struct {
		Receives struct {
			Prefixes []string
			Lengths  []int
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
	s.GenerateCall.Receives.Lengths = append(s.GenerateCall.Receives.Lengths, length)
	s.GenerateCall.Receives.Prefixes = append(s.GenerateCall.Receives.Prefixes, prefix)

	if s.GenerateCall.Stub != nil {
		return s.GenerateCall.Stub(prefix, length)
	}

	return s.GenerateCall.Returns.String, s.GenerateCall.Returns.Error
}
