package bosh

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type SSHKeyDeleter struct {
	stateStore stateStore
}

func NewSSHKeyDeleter(stateStore stateStore) SSHKeyDeleter {
	return SSHKeyDeleter{
		stateStore: stateStore,
	}
}

func (s SSHKeyDeleter) Delete() error {
	var err error
	varsDir, err := s.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	varsStore := filepath.Join(varsDir, "jumpbox-variables.yml")
	variables, err := ioutil.ReadFile(varsStore)
	if err == nil {
		varString, err := deleteJumpboxSSHKey(string(variables))
		if err != nil {
			return fmt.Errorf("Jumpbox variables: %s", err)
		}
		err = ioutil.WriteFile(varsStore, []byte(varString), os.ModePerm)
		if err != nil {
			return fmt.Errorf("Writing jumpbox vars store: %s", err) //not tested
		}
	}

	return nil
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
