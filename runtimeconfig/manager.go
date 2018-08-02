package runtimeconfig

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	logger             logger
	boshClientProvider boshClientProvider
	fs                 fs
	stateStore         stateStore
}

type logger interface {
	Step(string, ...interface{})
}

type stateStore interface {
	GetDirectorDeploymentDir() (string, error)
}

type fs interface {
	fileio.FileReader
}

type boshClientProvider interface {
	Client(jumpbox storage.Jumpbox, directorAddress, directorUsername, directorPassword, caCert string) (bosh.ConfigUpdater, error)
}

func NewManager(logger logger, boshClientProvider boshClientProvider, fs fs, stateStore stateStore) Manager {
	return Manager{
		logger:             logger,
		fs:                 fs,
		stateStore:         stateStore,
		boshClientProvider: boshClientProvider,
	}
}

func (m Manager) Update(state storage.State) error {
	boshClient, err := m.boshClientProvider.Client(state.Jumpbox, state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword, state.BOSH.DirectorSSLCA)
	if err != nil {
		return err // not tested
	}

	m.logger.Step("loading runtime config")
	yaml, err := m.LoadRuntimeConfigFile("dns.yml")
	if err != nil {
		return fmt.Errorf("could not open runtime config file %q: %s", "dns.yml", err)
	}

	m.logger.Step("applying runtime config")
	err = boshClient.UpdateRuntimeConfig(yaml, "dns")
	if err != nil {
		return err
	}

	return nil
}

func (m Manager) LoadRuntimeConfigFile(filename string) ([]byte, error) {
	dir, err := m.stateStore.GetDirectorDeploymentDir()
	if err != nil {
		return nil, err // untested
	}
	filePath := filepath.Join(dir, "runtime-configs", filename)
	bts, err := m.fs.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return bts, nil
}
