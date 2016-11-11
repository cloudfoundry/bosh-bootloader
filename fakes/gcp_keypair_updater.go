package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPKeyPairUpdater struct {
	UpdateCall struct {
		CallCount int
		Receives  struct {
			ProjectID string
		}
		Returns struct {
			KeyPair storage.KeyPair
			Error   error
		}
	}
}

func (g *GCPKeyPairUpdater) Update(projectID string) (storage.KeyPair, error) {
	g.UpdateCall.CallCount++
	g.UpdateCall.Receives.ProjectID = projectID

	return g.UpdateCall.Returns.KeyPair, g.UpdateCall.Returns.Error
}
