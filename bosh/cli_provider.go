package bosh

import (
	"io"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type CLIProvider struct {
	allProxyGetter allProxyGetter
	boshCLIPath    string
}

type allProxyGetter interface {
	GeneratePrivateKey() (string, error)
	BoshAllProxy(string, string) string
}

type AuthenticatedCLIRunner interface {
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

func NewCLIProvider(allProxyGetter allProxyGetter, boshCLIPath string) CLIProvider {
	return CLIProvider{
		allProxyGetter: allProxyGetter,
		boshCLIPath:    boshCLIPath,
	}
}

func (c CLIProvider) AuthenticatedCLI(jumpbox storage.Jumpbox, stderr io.Writer, directorAddress, directorUsername, directorPassword, directorCACert string) (AuthenticatedCLIRunner, error) {
	privateKey, err := c.allProxyGetter.GeneratePrivateKey()
	if err != nil {
		return AuthenticatedCLI{}, err
	}

	boshAllProxy := c.allProxyGetter.BoshAllProxy(jumpbox.URL, privateKey)
	return NewAuthenticatedCLI(stderr, c.boshCLIPath, directorAddress, directorUsername, directorPassword, directorCACert, boshAllProxy), nil
}
