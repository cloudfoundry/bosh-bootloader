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
	AddrCall struct {
		CallCount int
		Returns   struct {
			Addr string
		}
	}
}

func (s *Socks5Proxy) Start(jumpboxPrivateKey, jumpboxExternalURL string) error {
	s.StartCall.CallCount++
	s.StartCall.Receives.JumpboxPrivateKey = jumpboxPrivateKey
	s.StartCall.Receives.JumpboxExternalURL = jumpboxExternalURL

	return s.StartCall.Returns.Error
}

func (s *Socks5Proxy) Addr() string {
	s.AddrCall.CallCount++

	return s.AddrCall.Returns.Addr
}
