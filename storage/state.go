package storage

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

var encode func(io.Writer, interface{}) error = encodeFile

type KeyPair struct {
	Name       string `json:"name"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type AWS struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

type State struct {
	Version int      `json:"version"`
	AWS     AWS      `json:"aws"`
	KeyPair *KeyPair `json:"keyPair,omitempty"`
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
	state.Version = s.version

	file, err := os.OpenFile(filepath.Join(dir, "state.json"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	err = encode(file, state)
	if err != nil {
		return err
	}

	return nil
}

func (Store) Get(dir string) (State, error) {
	state := State{}
	file, err := os.OpenFile(filepath.Join(dir, "state.json"), os.O_RDONLY, os.ModePerm)
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
