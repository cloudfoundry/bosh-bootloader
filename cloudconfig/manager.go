package cloudconfig

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type fs interface {
	fileio.FileWriter
	fileio.DirReader
	fileio.Stater
}

type Manager struct {
	logger             logger
	command            command
	stateStore         stateStore
	opsGenerator       OpsGenerator
	boshClientProvider boshClientProvider
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

type boshClientProvider interface {
	Client(jumpbox storage.Jumpbox, directorAddress, directorUsername, directorPassword, caCert string) (bosh.ConfigUpdater, error)
}

type terraformManager interface {
	GetOutputs() (terraform.Outputs, error)
}

type stateStore interface {
	GetCloudConfigDir() (string, error)
	GetVarsDir() (string, error)
}

func NewManager(logger logger, cmd command, stateStore stateStore, opsGenerator OpsGenerator, boshClientProvider boshClientProvider,
	terraformManager terraformManager, fs fs) Manager {
	return Manager{
		logger:             logger,
		command:            cmd,
		stateStore:         stateStore,
		opsGenerator:       opsGenerator,
		boshClientProvider: boshClientProvider,
		terraformManager:   terraformManager,
		fs:                 fs,
	}
}

func (m Manager) Initialize(state storage.State) error {
	cloudConfigDir, err := m.stateStore.GetCloudConfigDir()
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

func (m Manager) GenerateVars(state storage.State) error {
	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return err
	}

	vars, err := m.opsGenerator.GenerateVars(state)
	if err != nil {
		return fmt.Errorf("Generate cloud config vars: %s", err)
	}

	err = m.fs.WriteFile(filepath.Join(varsDir, "cloud-config-vars.yml"), []byte(vars), storage.StateMode)
	if err != nil {
		return fmt.Errorf("Write cloud config vars: %s", err)
	}

	return nil
}

func (m Manager) IsPresentCloudConfig() bool {
	cloudConfigDir, err := m.stateStore.GetCloudConfigDir()
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
	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return false
	}

	_, err = m.fs.Stat(filepath.Join(varsDir, "cloud-config-vars.yml"))
	if err != nil {
		return false
	}

	return true
}

func (m Manager) Interpolate() (string, error) {
	cloudConfigDir, err := m.stateStore.GetCloudConfigDir()
	if err != nil {
		return "", err
	}

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return "", err
	}

	args := []string{
		"interpolate", filepath.Join(cloudConfigDir, "cloud-config.yml"),
		"--vars-file", filepath.Join(varsDir, "cloud-config-vars.yml"),
		"-o", filepath.Join(cloudConfigDir, "ops.yml"),
	}

	files, err := m.fs.ReadDir(cloudConfigDir)
	if err != nil {
		return "", fmt.Errorf("Read cloud config dir: %s", err)
	}

	for _, file := range files {
		name := file.Name()
		if name != "cloud-config.yml" && name != "ops.yml" {
			args = append(args, "-o", filepath.Join(cloudConfigDir, name))
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = m.command.Run(buf, cloudConfigDir, args)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (m Manager) Update(state storage.State) error {
	boshClient, err := m.boshClientProvider.Client(state.Jumpbox, state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword, state.BOSH.DirectorSSLCA)
	if err != nil {
		return err // not tested
	}

	m.logger.Step("generating cloud config")

	err = m.GenerateVars(state)
	if err != nil {
		return err
	}

	cloudConfig, err := m.Interpolate()
	if err != nil {
		return err
	}

	m.logger.Step("applying cloud config")
	err = boshClient.UpdateCloudConfig([]byte(cloudConfig))
	if err != nil {
		return err
	}

	return nil
}
