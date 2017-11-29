package bosh

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type SSHKeyGetter struct {
	stateStore stateStore
}

func NewSSHKeyGetter(stateStore stateStore) SSHKeyGetter {
	return SSHKeyGetter{
		stateStore: stateStore,
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
		return "", fmt.Errorf("Get vars directory: %s", err)
	}

	varsStore, err := ioutil.ReadFile(filepath.Join(varsDir, fmt.Sprintf("%s-vars-store.yml", deployment)))
	if err != nil {
		return "", fmt.Errorf("Read %s vars file: %s", deployment, err)
	}

	err = yaml.Unmarshal(varsStore, &p)
	if err != nil {
		return "", err
	}

	return p.JumpboxSSH.PrivateKey, nil
}
