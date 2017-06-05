package bosh

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	yaml "gopkg.in/yaml.v2"
)

type JumpboxSSHKeyGetter struct{}

type jumpboxVariables struct {
	JumpboxSSH struct {
		PrivateKey string `yaml:"private_key"`
	} `yaml:"jumpbox_ssh"`
}

func NewJumpboxSSHKeyGetter() JumpboxSSHKeyGetter {
	return JumpboxSSHKeyGetter{}
}

func (j JumpboxSSHKeyGetter) Get(state storage.State) (string, error) {
	var variables jumpboxVariables
	err := yaml.Unmarshal([]byte(state.Jumpbox.Variables), &variables)
	if err != nil {
		return "", err
	}

	return variables.JumpboxSSH.PrivateKey, nil
}
