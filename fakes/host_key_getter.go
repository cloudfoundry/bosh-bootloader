package fakes

import "golang.org/x/crypto/ssh"

type HostKeyGetter struct {
	GetCall struct {
		CallCount int
		Receives  struct {
			PrivateKey string
			ServerURL  string
		}
		Returns struct {
			HostKey ssh.PublicKey
			Error   error
		}
	}
}

func (h *HostKeyGetter) Get(privateKey, serverURL string) (ssh.PublicKey, error) {
	h.GetCall.CallCount++
	h.GetCall.Receives.PrivateKey = privateKey
	h.GetCall.Receives.ServerURL = serverURL

	return h.GetCall.Returns.HostKey, h.GetCall.Returns.Error
}
