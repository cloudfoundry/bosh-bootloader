package runtimeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	fileio.Remover
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

	if err := m.syncGCPLabelsOpsFile(dir, state); err != nil {
		return fmt.Errorf("failed to sync gcp labels ops file: %s", err)
	}

	opsFiles := []string{}
	files, err := m.fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read the runtime-config directory: %s", err)
	}

	for _, file := range files {
		name := file.Name()
		if name != "runtime-config.yml" && strings.HasSuffix(name, ".yml") {
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

// syncGCPLabelsOpsFile keeps the ops file that applies the GCP resource labels
// as runtime config tags in sync with the environment state. The director
// combines runtime config tags with the tags of every deployment, so all VMs
// deployed to the environment are labelled, not only the VMs that bbl creates
// itself. The file is removed when the environment has no GCP labels so that
// stale tags are not left behind.
func (m Manager) syncGCPLabelsOpsFile(dir string, state storage.State) error {
	path := filepath.Join(dir, "gcp-labels.yml")

	if state.IAAS != "gcp" || len(state.GCP.Labels) == 0 {
		if err := m.fs.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale ops file: %s", err)
		}
		return nil
	}

	contents, err := bosh.GCPLabelsRuntimeConfigOps(state.GCP.Labels)
	if err != nil {
		return fmt.Errorf("marshal ops file: %s", err)
	}

	if err := m.fs.WriteFile(path, contents, 0600); err != nil {
		return fmt.Errorf("write ops file: %s", err)
	}

	return nil
}
