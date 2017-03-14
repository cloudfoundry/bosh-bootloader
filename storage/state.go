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
)

type AWS struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

type GCP struct {
	ServiceAccountKey string `json:"serviceAccountKey"`
	ProjectID         string `json:"projectID"`
	Zone              string `json:"zone"`
	Region            string `json:"region"`
}

type Stack struct {
	Name            string `json:"name"`
	LBType          string `json:"lbType"`
	CertificateName string `json:"certificateName"`
	BOSHAZ          string `json:"boshAZ"`
}

type LB struct {
	Type   string `json:"type"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
	Domain string `json:"domain,omitempty"`
}

type State struct {
	Version    int     `json:"version"`
	IAAS       string  `json:"iaas"`
	NoDirector bool    `json:"noDirector"`
	AWS        AWS     `json:"aws,omitempty"`
	GCP        GCP     `json:"gcp,omitempty"`
	KeyPair    KeyPair `json:"keyPair,omitempty"`
	BOSH       BOSH    `json:"bosh,omitempty"`
	Stack      Stack   `json:"stack"`
	EnvID      string  `json:"envID"`
	TFState    string  `json:"tfState"`
	LB         LB      `json:"lb"`
}

type Store struct {
	version   int
	stateFile string
}

func NewStore(dir string) Store {
	return Store{
		version:   3,
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

func (g GCP) Empty() bool {
	return g.ServiceAccountKey == "" && g.ProjectID == "" && g.Region == "" && g.Zone == ""
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

	if state.Version < 3 {
		return state, errors.New("Existing bbl environment is incompatible with bbl v3. Create a new environment with v3 to continue.")
	}

	return state, nil
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
