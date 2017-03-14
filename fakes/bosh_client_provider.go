package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
)

type BOSHClientProvider struct {
	ClientCall struct {
		CallCount int

		Receives struct {
			DirectorAddress  string
			DirectorUsername string
			DirectorPassword string
		}
		Returns struct {
			Client bosh.Client
		}
	}
}

func (b *BOSHClientProvider) Client(directorAddress, directorUsername, directorPassword string) bosh.Client {
	b.ClientCall.CallCount++
	b.ClientCall.Receives.DirectorAddress = directorAddress
	b.ClientCall.Receives.DirectorUsername = directorUsername
	b.ClientCall.Receives.DirectorPassword = directorPassword
	return b.ClientCall.Returns.Client
}
