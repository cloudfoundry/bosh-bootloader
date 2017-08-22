package bosh

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	yaml "gopkg.in/yaml.v2"
)

type SSHKeyDeleter struct {
}

func NewSSHKeyDeleter() SSHKeyDeleter {
	return SSHKeyDeleter{}
}

func (SSHKeyDeleter) Delete(state storage.State) (storage.State, error) {
	var err error
	state.Jumpbox.Variables, err = deleteJumpboxSSHKey(state.Jumpbox.Variables)
	if err != nil {
		return storage.State{}, fmt.Errorf("Jumpbox variables: %s", err)
	}
	state.BOSH.Variables, err = deleteJumpboxSSHKey(state.BOSH.Variables)
	if err != nil {
		return storage.State{}, fmt.Errorf("BOSH variables: %s", err)
	}
	state.KeyPair = storage.KeyPair{}
	return state, nil
}

func deleteJumpboxSSHKey(varsString string) (string, error) {
	vars := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(varsString), &vars)
	if err != nil {
		return "", err
	}
	delete(vars, "jumpbox_ssh")
	newVars, err := yaml.Marshal(vars)
	if err != nil {
		return "", err // not tested
	}
	return string(newVars), nil
}
