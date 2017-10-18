package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	uuid "github.com/nu7hatch/gouuid"
)

var (
	marshalIndent = json.MarshalIndent
	uuidNewV4     = uuid.NewV4
)

const (
	STATE_VERSION = 12

	OS_READ_WRITE_MODE = os.FileMode(0644)
	StateFileName      = "bbl-state.json"
)

type Store struct {
	dir     string
	version int
}

func NewStore(dir string) Store {
	return Store{
		dir:     dir,
		version: STATE_VERSION,
	}
}

func (s Store) Set(state State) error {
	_, err := os.Stat(s.dir)
	if err != nil {
		return fmt.Errorf("Stat state dir: %s", err)
	}

	stateFile := filepath.Join(s.dir, StateFileName)
	if reflect.DeepEqual(state, State{}) {
		err := os.Remove(stateFile)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		rmdir := func(getDirFunc func() (string, error)) error {
			d, _ := getDirFunc()
			return os.RemoveAll(d)
		}
		if err := rmdir(s.GetBblDir); err != nil {
			return err
		}
		if err := rmdir(s.GetDirectorDeploymentDir); err != nil {
			return err
		}
		if err := rmdir(s.GetJumpboxDeploymentDir); err != nil {
			return err
		}
		if err := rmdir(s.GetVarsDir); err != nil {
			return err
		}
		if err := rmdir(s.GetTerraformDir); err != nil {
			return err
		}

		return nil
	}

	state.Version = s.version

	if state.ID == "" {
		uuid, err := uuidNewV4()
		if err != nil {
			return fmt.Errorf("Create state ID: %s", err)
		}
		state.ID = uuid.String()
	}

	state.AWS.AccessKeyID = ""
	state.AWS.SecretAccessKey = ""
	state.GCP.ServiceAccountKey = ""
	state.GCP.ProjectID = ""

	jsonData, err := marshalIndent(state, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(stateFile, jsonData, os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

func (s Store) GetStateDir() string {
	return s.dir
}

func (s Store) GetCloudConfigDir() (string, error) {
	return s.getDir(filepath.Join(".bbl", "cloudconfig"))
}

func (s Store) GetBblDir() (string, error) {
	return s.getDir(".bbl")
}

func (s Store) GetTerraformDir() (string, error) {
	return s.getDir("terraform")
}

func (s Store) GetVarsDir() (string, error) {
	return s.getDir("vars")
}

func (s Store) GetDirectorDeploymentDir() (string, error) {
	return s.getDir("bosh-deployment")
}

func (s Store) GetJumpboxDeploymentDir() (string, error) {
	return s.getDir("jumpbox-deployment")
}

func (s Store) getDir(name string) (string, error) {
	dir := filepath.Join(s.dir, name)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func stateAndBBLStateExist(dir string) (bool, error) {
	stateFile := filepath.Join(dir, "state.json")
	_, err := os.Stat(stateFile)
	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, err
	}

	bblStateFile := filepath.Join(dir, StateFileName)
	_, err = os.Stat(bblStateFile)
	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, err
	}
	return true, nil
}
