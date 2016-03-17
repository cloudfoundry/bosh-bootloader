package fakes

type StringGenerator struct {
	GenerateCall struct {
		Receives struct {
			Length int
		}
		Returns struct {
			String string
			Error  error
		}
	}
}

func (s *StringGenerator) Generate(length int) (string, error) {
	s.GenerateCall.Receives.Length = length

	return s.GenerateCall.Returns.String, s.GenerateCall.Returns.Error
}
