package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type BOSHClientProvider struct {
	ClientCall struct {
		CallCount int

		Receives struct {
			Jumpbox          storage.Jumpbox
			DirectorAddress  string
			DirectorUsername string
			DirectorPassword string
			DirectorCACert   string
		}
		Returns struct {
			Client bosh.Client
		}
	}
}

func (b *BOSHClientProvider) Client(jumpbox storage.Jumpbox, directorAddress, directorUsername, directorPassword, directorCACert string) (bosh.Client, error) {
	b.ClientCall.CallCount++
	b.ClientCall.Receives.Jumpbox = jumpbox
	b.ClientCall.Receives.DirectorAddress = directorAddress
	b.ClientCall.Receives.DirectorUsername = directorUsername
	b.ClientCall.Receives.DirectorPassword = directorPassword
	b.ClientCall.Receives.DirectorCACert = directorCACert
	return b.ClientCall.Returns.Client, nil
}
