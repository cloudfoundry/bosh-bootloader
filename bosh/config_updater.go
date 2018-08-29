package bosh

import (
	"fmt"
	"io"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type ConfigUpdater struct {
	boshCLIProvider boshCLIProvider
	boshCLI         AuthenticatedCLIRunner
}

type boshCLIProvider interface {
	AuthenticatedCLI(jumpbox storage.Jumpbox, stderr io.Writer, directorAddress, directorUsername, directorPassword, directorCACert string) (AuthenticatedCLIRunner, error)
}

func NewConfigUpdater(boshCLIProvider boshCLIProvider) ConfigUpdater {
	return ConfigUpdater{boshCLIProvider: boshCLIProvider}
}

func (c ConfigUpdater) InitializeAuthenticatedCLI(state storage.State) (AuthenticatedCLIRunner, error) {
	boshCLI, err := c.boshCLIProvider.AuthenticatedCLI(
		state.Jumpbox,
		os.Stderr,
		state.BOSH.DirectorAddress,
		state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword,
		state.BOSH.DirectorSSLCA,
	)

	if err != nil {
		return AuthenticatedCLI{}, fmt.Errorf("failed to create bosh cli: %s", err)
	}

	return boshCLI, nil
}

func (c ConfigUpdater) UpdateCloudConfig(boshCLI AuthenticatedCLIRunner, filepath string, opsFilepaths []string, varsFilepath string) error {
	args := []string{"update-cloud-config", filepath}
	for _, opsFilepath := range opsFilepaths {
		args = append(args, "--ops-file", opsFilepath)
	}
	args = append(args, "--vars-file", varsFilepath)

	return boshCLI.Run(nil, "", args)
}

func (c ConfigUpdater) UpdateRuntimeConfig(boshCLI AuthenticatedCLIRunner, filepath string, opsFilepaths []string, name string) error {
	args := []string{"update-runtime-config", filepath}
	for _, opsFilepath := range opsFilepaths {
		args = append(args, "--ops-file", opsFilepath)
	}
	args = append(args, "--name", name)

	return boshCLI.Run(nil, "", args)
}
