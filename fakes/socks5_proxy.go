package fakes

type Socks5Proxy struct {
	StartCall struct {
		CallCount int
		Receives  struct {
			Username    string
			PrivateKey  string
			ExternalURL string
		}
		Returns struct {
			Error error
		}
	}
	AddrCall struct {
		CallCount int
		Returns   struct {
			Addr  string
			Error error
		}
	}
}

func (s *Socks5Proxy) Start(username, privateKey, externalURL string) error {
	s.StartCall.CallCount++
	s.StartCall.Receives.Username = username
	s.StartCall.Receives.PrivateKey = privateKey
	s.StartCall.Receives.ExternalURL = externalURL

	return s.StartCall.Returns.Error
}

func (s *Socks5Proxy) Addr() (string, error) {
	s.AddrCall.CallCount++

	return s.AddrCall.Returns.Addr, s.AddrCall.Returns.Error
}
