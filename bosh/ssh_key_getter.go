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
	var vars string

	if state.Jumpbox.Enabled {
		vars = state.Jumpbox.Variables
	} else {
		vars = state.BOSH.Variables
	}

	if vars == "" {
		return state.KeyPair.PrivateKey, nil
	}

	var variables struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	err := yaml.Unmarshal([]byte(vars), &variables)
	if err != nil {
		return "", err
	}

	return variables.JumpboxSSH.PrivateKey, nil

}
