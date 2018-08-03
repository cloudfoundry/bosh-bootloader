package runtimeconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	logger             logger
	boshClientProvider boshClientProvider
	dirProvider        dirProvider
}

type logger interface {
	Step(string, ...interface{})
}

type boshClientProvider interface {
	BoshCLI(jumpbox storage.Jumpbox, stderr io.Writer, directorAddress, directorUsername, directorPassword, directorCACert string) (bosh.RuntimeConfigUpdater, error)
}

type dirProvider interface {
	GetDirectorDeploymentDir() (string, error)
}

func NewManager(logger logger, dirProvider dirProvider, boshClientProvider boshClientProvider) Manager {
	return Manager{
		logger:             logger,
		boshClientProvider: boshClientProvider,
		dirProvider:        dirProvider,
	}
}

func (m Manager) Update(state storage.State) error {
	boshCLI, err := m.boshClientProvider.BoshCLI(state.Jumpbox,
		os.Stderr,
		state.BOSH.DirectorAddress,
		state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword,
		state.BOSH.DirectorSSLCA,
	)
	if err != nil {
		panic(err)
	}

	dir, err := m.dirProvider.GetDirectorDeploymentDir()
	if err != nil {
		return fmt.Errorf("could not find bosh-deployment directory: %s", err)
	}

	filename := "dns.yml"
	m.logger.Step("applying runtime config")
	filepath := filepath.Join(dir, "runtime-configs", filename)

	err = boshCLI.UpdateRuntimeConfig(filepath, "dns")
	if err != nil {
		return fmt.Errorf("failed to update runtime-config: %s", err)
	}

	return nil
}
