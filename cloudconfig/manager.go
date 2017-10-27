package cloudconfig

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

var (
	tempDir   func(string, string) (string, error)    = ioutil.TempDir
	writeFile func(string, []byte, os.FileMode) error = ioutil.WriteFile
)

type Manager struct {
	logger             logger
	command            command
	stateStore         stateStore
	opsGenerator       OpsGenerator
	boshClientProvider boshClientProvider
	terraformManager   terraformManager
	sshKeyGetter       sshKeyGetter
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
	Client(jumpbox storage.Jumpbox, directorAddress, directorUsername, directorPassword, caCert string) (bosh.Client, error)
}

type terraformManager interface {
	GetOutputs(storage.State) (terraform.Outputs, error)
}

type sshKeyGetter interface {
	Get(string) (string, error)
}

type stateStore interface {
	GetCloudConfigDir() (string, error)
	GetVarsDir() (string, error)
}

func NewManager(logger logger, cmd command, stateStore stateStore, opsGenerator OpsGenerator, boshClientProvider boshClientProvider,
	terraformManager terraformManager, sshKeyGetter sshKeyGetter) Manager {
	return Manager{
		logger:             logger,
		command:            cmd,
		stateStore:         stateStore,
		opsGenerator:       opsGenerator,
		boshClientProvider: boshClientProvider,
		terraformManager:   terraformManager,
		sshKeyGetter:       sshKeyGetter,
	}
}

func (m Manager) Generate(state storage.State) (string, error) {
	cloudConfigDir, err := m.stateStore.GetCloudConfigDir()
	if err != nil {
		return "", fmt.Errorf("Get cloud config dir: %s", err)
	}

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars dir: %s", err)
	}

	err = writeFile(filepath.Join(cloudConfigDir, "cloud-config.yml"), []byte(BaseCloudConfig), os.ModePerm)
	if err != nil {
		return "", err
	}

	ops, err := m.opsGenerator.Generate(state)
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(cloudConfigDir, "ops.yml"), []byte(ops), os.ModePerm)
	if err != nil {
		return "", err
	}

	vars, err := m.opsGenerator.GenerateVars(state)
	if err != nil {
		return "", fmt.Errorf("Generate cloud config vars: %s", err)
	}

	err = writeFile(filepath.Join(varsDir, "cloud-config-vars.yml"), []byte(vars), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Write cloud config vars: %s", err)
	}

	args := []string{
		"interpolate", filepath.Join(cloudConfigDir, "cloud-config.yml"),
		"-o", filepath.Join(cloudConfigDir, "ops.yml"),
		"--vars-file", filepath.Join(varsDir, "cloud-config-vars.yml"),
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
	cloudConfig, err := m.Generate(state)
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
