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
	GetPasswordCall struct {
		CallCount int
		Returns   struct {
			Password string
			Error    error
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

func (s *CredhubGetter) GetPassword() (string, error) {
	s.GetPasswordCall.CallCount++

	return s.GetPasswordCall.Returns.Password, s.GetPasswordCall.Returns.Error
}
