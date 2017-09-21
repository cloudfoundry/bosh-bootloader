package bosh

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	yaml "gopkg.in/yaml.v2"
)

type SSHKeyGetter struct{}

func NewSSHKeyGetter() SSHKeyGetter {
	return SSHKeyGetter{}
}

func (j SSHKeyGetter) Get(state storage.State) (string, error) {
	var variables struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	vars := state.Jumpbox.Variables
	err := yaml.Unmarshal([]byte(vars), &variables)
	if err != nil {
		return "", err
	}

	return variables.JumpboxSSH.PrivateKey, nil
}
