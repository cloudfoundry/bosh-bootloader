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
	logger             logger
	command            command
	opsGenerator       opsGenerator
	boshClientProvider boshClientProvider
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

func NewManager(logger logger, cmd command, opsGenerator opsGenerator, boshClientProvider boshClientProvider) Manager {
	return Manager{
		logger:             logger,
		command:            cmd,
		opsGenerator:       opsGenerator,
		boshClientProvider: boshClientProvider,
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
