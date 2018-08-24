package runtimeconfig

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	logger             logger
	boshClientProvider boshClientProvider
	dirProvider        dirProvider
	cli                runner
	fs                 fs
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

type runner interface {
	Run(stdout io.Writer, runtimeConfigDirectory string, args []string) error
}

type boshClientProvider interface {
	BoshCLI(jumpbox storage.Jumpbox, stderr io.Writer, directorAddress, directorUsername, directorPassword, directorCACert string) (bosh.RuntimeConfigUpdater, error)
}

type dirProvider interface {
	GetDirectorDeploymentDir() (string, error)
	GetRuntimeConfigDir() (string, error)
}

func NewManager(logger logger, dirProvider dirProvider, boshClientProvider boshClientProvider, cli runner, fs fs) Manager {
	return Manager{
		logger:             logger,
		boshClientProvider: boshClientProvider,
		dirProvider:        dirProvider,
		cli:                cli,
		fs:                 fs,
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

func (m Manager) Interpolate() (string, error) {
	runtimeConfigDir, err := m.dirProvider.GetRuntimeConfigDir()
	if err != nil {
		panic(err)
	}

	args := []string{
		"interpolate", filepath.Join(runtimeConfigDir, "runtime-config.yml"),
	}

	files, err := m.fs.ReadDir(runtimeConfigDir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		name := file.Name()
		if name != "runtime-config.yml" {
			args = append(args, "-o", filepath.Join(runtimeConfigDir, name))
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = m.cli.Run(buf, runtimeConfigDir, args)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
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

	err = boshCLI.UpdateRuntimeConfig(dir, "dns")
	if err != nil {
		return fmt.Errorf("failed to update runtime-config: %s", err)
	}

	return nil
}
