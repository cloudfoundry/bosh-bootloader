package fakes

type Socks5Proxy struct {
	StartCall struct {
		CallCount int
		Receives  struct {
			JumpboxPrivateKey  string
			JumpboxExternalURL string
		}
		Returns struct {
			Error error
		}
	}
	StopCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (s *Socks5Proxy) Start(jumpboxPrivateKey, jumpboxExternalURL string) error {
	s.StartCall.CallCount++
	s.StartCall.Receives.JumpboxPrivateKey = jumpboxPrivateKey
	s.StartCall.Receives.JumpboxExternalURL = jumpboxExternalURL

	return s.StartCall.Returns.Error
}

func (s *Socks5Proxy) Stop() error {
	s.StopCall.CallCount++

	return s.StopCall.Returns.Error
}
