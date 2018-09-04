package runtimeconfig

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	logger               logger
	runtimeConfigUpdater configUpdater
	dirProvider          dirProvider
	fs                   fs
}

type fs interface {
	fileio.FileWriter
	fileio.DirReader
	fileio.Stater
	fileio.FileReader
}

type logger interface {
	Step(string, ...interface{})
}

type dirProvider interface {
	GetDirectorDeploymentDir() (string, error)
	GetRuntimeConfigDir() (string, error)
}

type configUpdater interface {
	InitializeAuthenticatedCLI(state storage.State) (bosh.AuthenticatedCLIRunner, error)
	UpdateRuntimeConfig(boshCLI bosh.AuthenticatedCLIRunner, filepath string, opsFilepaths []string, name string) error
}

func NewManager(logger logger, dirProvider dirProvider, runtimeConfigUpdater configUpdater, fs fs) Manager {
	return Manager{
		logger:               logger,
		runtimeConfigUpdater: runtimeConfigUpdater,
		dirProvider:          dirProvider,
		fs:                   fs,
	}
}

func (m Manager) Initialize(state storage.State) error {
	runtimeConfigsDir, err := m.dirProvider.GetRuntimeConfigDir()
	if err != nil {
		return fmt.Errorf("runtime config directory could not be found: %s", err)
	}

	directorDir, err := m.dirProvider.GetDirectorDeploymentDir()
	if err != nil {
		return fmt.Errorf("bosh-deployment directory could not be found: %s", err)
	}

	path := filepath.Join(directorDir, "runtime-configs", "dns.yml")

	buf, err := m.fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read runtime config dns.yml from bosh-deployment: %s", err)
	}
	err = m.fs.WriteFile(filepath.Join(runtimeConfigsDir, "runtime-config.yml"), buf, 0600)
	if err != nil {
		return fmt.Errorf("failed to write runtime config: %s", err)
	}

	return nil
}

func (m Manager) Update(state storage.State) error {
	boshCLI, err := m.runtimeConfigUpdater.InitializeAuthenticatedCLI(state)
	if err != nil {
		return fmt.Errorf("failed to initialize authenticated bosh cli: %s", err)
	}

	dir, err := m.dirProvider.GetRuntimeConfigDir()
	if err != nil {
		return fmt.Errorf("could not find runtime-config directory: %s", err)
	}

	opsFiles := []string{}
	files, err := m.fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read the runtime-config directory: %s", err)
	}

	for _, file := range files {
		name := file.Name()
		if name != "runtime-config.yml" {
			opsFiles = append(opsFiles, filepath.Join(dir, name))
		}
	}

	m.logger.Step("applying runtime config")
	runtimeConfigPath := filepath.Join(dir, "runtime-config.yml")
	err = m.runtimeConfigUpdater.UpdateRuntimeConfig(boshCLI, runtimeConfigPath, opsFiles, "dns")
	if err != nil {
		return fmt.Errorf("failed to update runtime-config: %s", err)
	}

	return nil
}
