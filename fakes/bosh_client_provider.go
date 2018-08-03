package fakes

import (
	"io"

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
			Client bosh.ConfigUpdater
			Error  error
		}
	}
	BoshCLICall struct {
		CallCount int

		Receives struct {
			Jumpbox          storage.Jumpbox
			Stderr           io.Writer
			DirectorAddress  string
			DirectorUsername string
			DirectorPassword string
			DirectorCACert   string
		}
		Returns struct {
			BoshCLI bosh.RuntimeConfigUpdater
			Error   error
		}
	}
}

func (b *BOSHClientProvider) Client(jumpbox storage.Jumpbox, directorAddress, directorUsername, directorPassword, directorCACert string) (bosh.ConfigUpdater, error) {
	b.ClientCall.CallCount++
	b.ClientCall.Receives.Jumpbox = jumpbox
	b.ClientCall.Receives.DirectorAddress = directorAddress
	b.ClientCall.Receives.DirectorUsername = directorUsername
	b.ClientCall.Receives.DirectorPassword = directorPassword
	b.ClientCall.Receives.DirectorCACert = directorCACert
	return b.ClientCall.Returns.Client, b.ClientCall.Returns.Error
}

func (b *BOSHClientProvider) BoshCLI(jumpbox storage.Jumpbox, stderr io.Writer, directorAddress, directorUsername, directorPassword, directorCACert string) (bosh.RuntimeConfigUpdater, error) {
	b.BoshCLICall.CallCount++
	b.BoshCLICall.Receives.Jumpbox = jumpbox
	b.BoshCLICall.Receives.Stderr = stderr
	b.BoshCLICall.Receives.DirectorAddress = directorAddress
	b.BoshCLICall.Receives.DirectorUsername = directorUsername
	b.BoshCLICall.Receives.DirectorPassword = directorPassword
	b.BoshCLICall.Receives.DirectorCACert = directorCACert
	return b.BoshCLICall.Returns.BoshCLI, b.BoshCLICall.Returns.Error

}
