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
	Run(stdout io.Writer, cloudConfigDirectory string, args []string) error
}

type boshClientProvider interface {
	BoshCLI(jumpbox storage.Jumpbox, stderr io.Writer, directorAddress, directorUsername, directorPassword, directorCACert string) (bosh.RuntimeConfigUpdater, error)
}

type dirProvider interface {
	GetDirectorDeploymentDir() (string, error)
	GetRuntimeConfigsDir() (string, error)
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
	runtimeConfigsDir, err := m.dirProvider.GetRuntimeConfigsDir()
	if err != nil {
		panic(err)
	}

	directorDir, err := m.dirProvider.GetDirectorDeploymentDir()
	if err != nil {
		panic(err)
	}

	filename := "dns.yml"
	path := filepath.Join(directorDir, "runtime-configs", filename)
	
	dnsConfig, err := m.fs.ReadFile(path)
	if err != nil {
		panic(err)
	}


	return nil
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

	err = boshCLI.UpdateRuntimeConfig(filepath, "dns")
	if err != nil {
		return fmt.Errorf("failed to update runtime-config: %s", err)
	}

	return nil
}

func (m Manager) Interpolate() (string, error) {
	dir, err := m.dirProvider.GetDirectorDeploymentDir()
	if err != nil {
		return "", fmt.Errorf("could not find bosh-deployment directory: %s", err)
	}
	buf := bytes.NewBuffer([]byte{})
	args := []string{
		"interpolate",
		filepath.Join(dir, "runtime-configs"),
		// ??? // "-o", filepath.Join(dir, "ops.yml")
	}
	err = m.cli.Run(buf, "", args)
	return "", err
}
