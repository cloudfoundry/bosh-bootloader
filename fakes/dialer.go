package fakes

import "net"

type Dialer struct {
	DialCall struct {
		CallCount int
		Stub      func(network, addr string) (net.Conn, error)
		Receives  struct {
			Network string
			Addr    string
		}
		Returns struct {
			Connection net.Conn
			Error      error
		}
	}
}

func (s *Dialer) Dial(network, addr string) (net.Conn, error) {
	s.DialCall.CallCount++
	s.DialCall.Receives.Network = network
	s.DialCall.Receives.Addr = addr

	if s.DialCall.Stub != nil {
		return s.DialCall.Stub(network, addr)
	}

	return s.DialCall.Returns.Connection, s.DialCall.Returns.Error
}
