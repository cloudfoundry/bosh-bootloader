package fakes

import (
	"io"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type BOSHCLIProvider struct {
	AuthenticatedCLICall struct {
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
			AuthenticatedCLI bosh.AuthenticatedCLIRunner
			Error            error
		}
	}
}

func (b *BOSHCLIProvider) AuthenticatedCLI(jumpbox storage.Jumpbox, stderr io.Writer, directorAddress, directorUsername, directorPassword, directorCACert string) (bosh.AuthenticatedCLIRunner, error) {
	b.AuthenticatedCLICall.CallCount++
	b.AuthenticatedCLICall.Receives.Jumpbox = jumpbox
	b.AuthenticatedCLICall.Receives.Stderr = stderr
	b.AuthenticatedCLICall.Receives.DirectorAddress = directorAddress
	b.AuthenticatedCLICall.Receives.DirectorUsername = directorUsername
	b.AuthenticatedCLICall.Receives.DirectorPassword = directorPassword
	b.AuthenticatedCLICall.Receives.DirectorCACert = directorCACert
	return b.AuthenticatedCLICall.Returns.AuthenticatedCLI, b.AuthenticatedCLICall.Returns.Error

}
