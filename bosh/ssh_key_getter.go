package bosh

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fileio"

	"gopkg.in/yaml.v2"
)

type SSHKeyGetter struct {
	stateStore stateStore
	fReader    fileio.FileReader
}

func NewSSHKeyGetter(stateStore stateStore, fReader fileio.FileReader) SSHKeyGetter {
	return SSHKeyGetter{
		stateStore: stateStore,
		fReader:    fReader,
	}
}

func (j SSHKeyGetter) Get(deployment string) (string, error) {
	var p struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	varsDir, err := j.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars directory: %s", err) //nolint:staticcheck
	}

	varsStore, err := j.fReader.ReadFile(filepath.Join(varsDir, fmt.Sprintf("%s-vars-store.yml", deployment)))
	if err != nil {
		return "", fmt.Errorf("Read %s vars file: %s", deployment, err) //nolint:staticcheck
	}

	err = yaml.Unmarshal(varsStore, &p)
	if err != nil {
		return "", err
	}

	return p.JumpboxSSH.PrivateKey, nil
}
