package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPKeyPairUpdater struct {
	UpdateCall struct {
		CallCount int
		Returns   struct {
			KeyPair storage.KeyPair
			Error   error
		}
	}
}

func (g *GCPKeyPairUpdater) Update() (storage.KeyPair, error) {
	g.UpdateCall.CallCount++

	return g.UpdateCall.Returns.KeyPair, g.UpdateCall.Returns.Error
}
