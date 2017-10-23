package fakes

type SSHKeyGetter struct {
	GetCall struct {
		CallCount int
		Receives  struct {
			Variables string
		}
		Returns struct {
			PrivateKey string
			Error      error
		}
	}
}

func (s *SSHKeyGetter) Get(variables string) (string, error) {
	s.GetCall.CallCount++
	s.GetCall.Receives.Variables = variables

	return s.GetCall.Returns.PrivateKey, s.GetCall.Returns.Error
}
