package cloudconfig

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type configUpdater interface {
	InitializeAuthenticatedCLI(state storage.State) (bosh.AuthenticatedCLIRunner, error)
	UpdateCloudConfig(boshCLI bosh.AuthenticatedCLIRunner, filepath string, opsFilepaths []string, varsFilepath string) error
}

type fs interface {
	fileio.FileWriter
	fileio.DirReader
	fileio.Stater
}

type Manager struct {
	logger             logger
	cloudConfigUpdater configUpdater
	dirProvider        dirProvider
	opsGenerator       OpsGenerator
	terraformManager   terraformManager
	fs                 fs
}

type logger interface {
	Step(string, ...interface{})
}

type command interface {
	Run(stdout io.Writer, cloudConfigDirectory string, args []string) error
}

type OpsGenerator interface {
	Generate(state storage.State) (string, error)
	GenerateVars(state storage.State) (string, error)
}

type boshCLIProvider interface {
	Client(jumpbox storage.Jumpbox, directorAddress, directorUsername, directorPassword, caCert string) (configUpdater, error)
}

type terraformManager interface {
	GetOutputs() (terraform.Outputs, error)
}

type dirProvider interface {
	GetCloudConfigDir() (string, error)
	GetVarsDir() (string, error)
}

func NewManager(logger logger, cloudConfigUpdater configUpdater, dirProvider dirProvider, opsGenerator OpsGenerator,
	terraformManager terraformManager, fs fs) Manager {
	return Manager{
		logger:             logger,
		cloudConfigUpdater: cloudConfigUpdater,
		dirProvider:        dirProvider,
		opsGenerator:       opsGenerator,
		terraformManager:   terraformManager,
		fs:                 fs,
	}
}

func (m Manager) Initialize(state storage.State) error {
	cloudConfigDir, err := m.dirProvider.GetCloudConfigDir()
	if err != nil {
		return err
	}

	err = m.fs.WriteFile(filepath.Join(cloudConfigDir, "cloud-config.yml"), []byte(BaseCloudConfig), storage.StateMode)
	if err != nil {
		return err
	}

	ops, err := m.opsGenerator.Generate(state)
	if err != nil {
		return err
	}

	err = m.fs.WriteFile(filepath.Join(cloudConfigDir, "ops.yml"), []byte(ops), storage.StateMode)
	if err != nil {
		return err
	}

	return nil
}

func (m Manager) IsPresentCloudConfig() bool {
	cloudConfigDir, err := m.dirProvider.GetCloudConfigDir()
	if err != nil {
		return false
	}

	_, err1 := m.fs.Stat(filepath.Join(cloudConfigDir, "cloud-config.yml"))
	_, err2 := m.fs.Stat(filepath.Join(cloudConfigDir, "ops.yml"))
	if err1 != nil || err2 != nil {
		return false
	}

	return true
}

func (m Manager) IsPresentCloudConfigVars() bool {
	varsDir, err := m.dirProvider.GetVarsDir()
	if err != nil {
		return false
	}

	_, err = m.fs.Stat(filepath.Join(varsDir, "cloud-config-vars.yml"))
	if err != nil {
		return false
	}

	return true
}

func (m Manager) Update(state storage.State) error {
	boshCLI, err := m.cloudConfigUpdater.InitializeAuthenticatedCLI(state)
	if err != nil {
		return fmt.Errorf("failed to initialize authenticated bosh cli: %s", err)
	}

	m.logger.Step("generating cloud config")

	varsDir, err := m.dirProvider.GetVarsDir()
	if err != nil {
		return fmt.Errorf("could not find vars directory: %s", err)
	}
	vars, err := m.opsGenerator.GenerateVars(state)
	if err != nil {
		return fmt.Errorf("failed to generate cloud config vars: %s", err)
	}

	err = m.fs.WriteFile(filepath.Join(varsDir, "cloud-config-vars.yml"), []byte(vars), storage.StateMode)
	if err != nil {
		return fmt.Errorf("failed to write cloud config vars: %s", err)
	}

	cloudConfigDir, err := m.dirProvider.GetCloudConfigDir()
	if err != nil {
		return fmt.Errorf("could not find cloud-config directory: %s", err)
	}
	cloudConfigPath := filepath.Join(cloudConfigDir, "cloud-config.yml")

	varsFilepath := filepath.Join(varsDir, "cloud-config-vars.yml")

	opsFiles := []string{filepath.Join(cloudConfigDir, "ops.yml")}
	files, err := m.fs.ReadDir(cloudConfigDir)
	if err != nil {
		return fmt.Errorf("failed to read the cloud-config directory: %s", err)
	}

	for _, file := range files {
		name := file.Name()
		if name != "cloud-config.yml" && name != "ops.yml" {
			opsFiles = append(opsFiles, filepath.Join(cloudConfigDir, name))
		}
	}

	m.logger.Step("applying cloud config")
	err = m.cloudConfigUpdater.UpdateCloudConfig(boshCLI, cloudConfigPath, opsFiles, varsFilepath)
	if err != nil {
		return fmt.Errorf("failed to update cloud-config: %s", err)
	}

	return nil
}
