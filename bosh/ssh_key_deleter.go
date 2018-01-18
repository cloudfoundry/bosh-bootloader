package bosh

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	yaml "gopkg.in/yaml.v2"
)

type fs interface {
	fileio.FileReader
	fileio.FileWriter
}

type SSHKeyDeleter struct {
	stateStore stateStore
	fs         fs
}

func NewSSHKeyDeleter(stateStore stateStore, fs fs) SSHKeyDeleter {
	return SSHKeyDeleter{
		stateStore: stateStore,
		fs:         fs,
	}
}

func (s SSHKeyDeleter) Delete() error {
	var err error
	varsDir, err := s.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	varsStore := filepath.Join(varsDir, "jumpbox-vars-store.yml")
	variables, err := s.fs.ReadFile(varsStore)
	if err == nil {
		varString, err := deleteJumpboxSSHKey(string(variables))
		if err != nil {
			return fmt.Errorf("Jumpbox variables: %s", err)
		}
		if string(variables) == varString {
			return nil
		}
		err = s.fs.WriteFile(varsStore, []byte(varString), storage.StateMode)
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
