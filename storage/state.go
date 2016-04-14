package storage

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

const OS_READ_WRITE_MODE = os.FileMode(0644)

var encode func(io.Writer, interface{}) error = encodeFile

type AWS struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

type Stack struct {
	Name   string `json:"name"`
	LBType string `json:"lbType"`
}

type State struct {
	Version int     `json:"version"`
	AWS     AWS     `json:"aws"`
	KeyPair KeyPair `json:"keyPair,omitempty"`
	BOSH    BOSH    `json:"bosh,omitempty"`
	Stack   Stack   `json:"stack"`
}

type Store struct {
	version int
}

func NewStore() Store {
	return Store{
		version: 1,
	}
}

func (s Store) Set(dir string, state State) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(state, State{}) {
		err := os.Remove(filepath.Join(dir, "state.json"))
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		return nil
	}

	file, err := os.OpenFile(filepath.Join(dir, "state.json"), os.O_RDWR|os.O_CREATE, OS_READ_WRITE_MODE)
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

func (Store) Get(dir string) (State, error) {
	state := State{}

	_, err := os.Stat(dir)
	if err != nil {
		return state, err
	}

	file, err := os.Open(filepath.Join(dir, "state.json"))
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

func encodeFile(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
