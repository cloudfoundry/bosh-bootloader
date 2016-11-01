package storage

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

const (
	OS_READ_WRITE_MODE = os.FileMode(0644)
	StateFileName      = "bbl-state.json"
)

type logger interface {
	Println(message string)
}

var (
	encode func(io.Writer, interface{}) error = encodeFile
	rename func(string, string) error         = os.Rename
)

type AWS struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

type Stack struct {
	Name            string `json:"name"`
	LBType          string `json:"lbType"`
	CertificateName string `json:"certificateName"`
}

type State struct {
	Version int     `json:"version"`
	IAAS    string  `json:"iaas"`
	AWS     AWS     `json:"aws"`
	KeyPair KeyPair `json:"keyPair,omitempty"`
	BOSH    BOSH    `json:"bosh,omitempty"`
	Stack   Stack   `json:"stack"`
	EnvID   string  `json:"envID"`
}

type Store struct {
	version   int
	stateFile string
}

func NewStore(dir string) Store {
	return Store{
		version:   1,
		stateFile: filepath.Join(dir, StateFileName),
	}
}

func (s Store) Set(state State) error {
	_, err := os.Stat(filepath.Dir(s.stateFile))
	if err != nil {
		return err
	}

	if reflect.DeepEqual(state, State{}) {
		err := os.Remove(s.stateFile)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		return nil
	}

	file, err := os.OpenFile(s.stateFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, OS_READ_WRITE_MODE)
	if err != nil {
		return err
	}

	state.Version = s.version
	err = encode(file, state)
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

	bothExist, err := stateAndBBLStateExist(dir)
	if err != nil {
		return state, err
	}

	if bothExist {
		return state, errors.New("Cannot proceed with state.json and bbl-state.json present. Please delete one of the files.")
	}

	err = renameStateToBBLState(dir)
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

	return state, nil
}

func renameStateToBBLState(dir string) error {
	stateFile := filepath.Join(dir, "state.json")
	_, err := os.Stat(stateFile)
	switch {
	case os.IsNotExist(err):
		return nil
	case err == nil:
		GetStateLogger.Println("renaming state.json to bbl-state.json")
		err := rename(stateFile, filepath.Join(dir, StateFileName))
		if err != nil {
			return err
		}
		return nil
	default:
		return err
	}
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

func encodeFile(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
