package bosh

import (
	yaml "gopkg.in/yaml.v2"
)

type SSHKeyGetter struct{}

func NewSSHKeyGetter() SSHKeyGetter {
	return SSHKeyGetter{}
}

func (j SSHKeyGetter) Get(vars string) (string, error) {
	var p struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	err := yaml.Unmarshal([]byte(vars), &p)
	if err != nil {
		return "", err
	}

	return p.JumpboxSSH.PrivateKey, nil
}
