package storage

import (
	"encoding/json"
	"errors"
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

type logger interface {
	Println(message string)
}

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

var GetStateLogger logger

func GetState(dir string) (State, error) {
	state := State{}

	_, err := os.Stat(dir)
	if err != nil {
		return state, err
	}

	file, err := os.Open(filepath.Join(dir, StateFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, err
	}

	err = json.NewDecoder(file).Decode(&state)
	if err != nil {
		return state, err
	}

	emptyState := State{}
	if reflect.DeepEqual(state, emptyState) {
		state = State{
			Version: STATE_VERSION,
		}
	}

	if state.Version < 3 {
		return state, errors.New("Existing bbl environment is incompatible with bbl v3. Create a new environment with v3 to continue.")
	}

	if state.Version > STATE_VERSION {
		return state, fmt.Errorf("Existing bbl environment was created with a newer version of bbl. Please upgrade to a version of bbl compatible with schema version %d.\n", state.Version)
	}

	if state.Version < STATE_VERSION {
		GetStateLogger.Println(fmt.Sprintf("Warning: Current schema version (%d) is newer than existing bbl environment schema (%d). Some things may not work as expected until you bbl up again.", STATE_VERSION, state.Version))
	}

	return state, nil
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
