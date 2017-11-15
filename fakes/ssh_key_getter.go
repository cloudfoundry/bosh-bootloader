package fakes

type SSHKeyGetter struct {
	GetCall struct {
		CallCount int
		Receives  struct {
			Deployment string
		}
		Returns struct {
			PrivateKey string
			Error      error
		}
	}
}

func (s *SSHKeyGetter) Get(deployment string) (string, error) {
	s.GetCall.CallCount++
	s.GetCall.Receives.Deployment = deployment

	return s.GetCall.Returns.PrivateKey, s.GetCall.Returns.Error
}
