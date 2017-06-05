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
)

var (
	tempDir   func(string, string) (string, error)    = ioutil.TempDir
	writeFile func(string, []byte, os.FileMode) error = ioutil.WriteFile
)

type Manager struct {
	logger              logger
	command             command
	opsGenerator        opsGenerator
	boshClientProvider  boshClientProvider
	socks5Proxy         socks5Proxy
	terraformManager    terraformManager
	jumpboxSSHKeyGetter jumpboxSSHKeyGetter
}

type logger interface {
	Step(string, ...interface{})
}

type command interface {
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

type opsGenerator interface {
	Generate(state storage.State) (string, error)
}

type boshClientProvider interface {
	Client(directorAddress, directorUsername, directorPassword string) bosh.Client
}

type socks5Proxy interface {
	Start(string, string) error
}

type terraformManager interface {
	GetOutputs(storage.State) (map[string]interface{}, error)
}

type jumpboxSSHKeyGetter interface {
	Get(storage.State) (string, error)
}

func NewManager(logger logger, cmd command, opsGenerator opsGenerator, boshClientProvider boshClientProvider,
	socks5Proxy socks5Proxy, terraformManager terraformManager, jumpboxSSHKeyGetter jumpboxSSHKeyGetter) Manager {
	return Manager{
		logger:              logger,
		command:             cmd,
		opsGenerator:        opsGenerator,
		boshClientProvider:  boshClientProvider,
		socks5Proxy:         socks5Proxy,
		terraformManager:    terraformManager,
		jumpboxSSHKeyGetter: jumpboxSSHKeyGetter,
	}
}

func (m Manager) Generate(state storage.State) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	workingDir, err := tempDir("", "")
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(workingDir, "cloud-config.yml"), []byte(BaseCloudConfig), os.ModePerm)
	if err != nil {
		return "", err
	}

	ops, err := m.opsGenerator.Generate(state)
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(workingDir, "ops.yml"), []byte(ops), os.ModePerm)
	if err != nil {
		return "", err
	}

	args := []string{
		"interpolate", fmt.Sprintf("%s/cloud-config.yml", workingDir),
		"-o", fmt.Sprintf("%s/ops.yml", workingDir),
	}

	err = m.command.Run(buf, workingDir, args)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (m Manager) Update(state storage.State) error {
	if state.Jumpbox.Enabled {
		privateKey, err := m.jumpboxSSHKeyGetter.Get(state)
		if err != nil {
			return err
		}
		terraformOutputs, err := m.terraformManager.GetOutputs(state)
		if err != nil {
			return err
		}
		jumpboxURL := fmt.Sprintf("%s:%d", terraformOutputs["external_ip"], 22)

		m.logger.Step("starting socks5 proxy")
		err = m.socks5Proxy.Start(privateKey, jumpboxURL)
		if err != nil {
			return err
		}
	}

	m.logger.Step("generating cloud config")
	cloudConfig, err := m.Generate(state)
	if err != nil {
		return err
	}

	m.logger.Step("applying cloud config")
	boshClient := m.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)
	err = boshClient.UpdateCloudConfig([]byte(cloudConfig))
	if err != nil {
		return err
	}

	return nil
}
