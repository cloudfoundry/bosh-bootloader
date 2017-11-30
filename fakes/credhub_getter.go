package fakes

type CredhubGetter struct {
	GetServerCall struct {
		CallCount int
		Returns   struct {
			Server string
			Error  error
		}
	}
	GetCertsCall struct {
		CallCount int
		Returns   struct {
			Certs string
			Error error
		}
	}
}

func (s *CredhubGetter) GetServer() (string, error) {
	s.GetServerCall.CallCount++

	return s.GetServerCall.Returns.Server, s.GetServerCall.Returns.Error
}

func (s *CredhubGetter) GetCerts() (string, error) {
	s.GetCertsCall.CallCount++

	return s.GetCertsCall.Returns.Certs, s.GetCertsCall.Returns.Error
}
