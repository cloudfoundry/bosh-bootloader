package state

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

var encode func(io.Writer, interface{}) error = encodeFile

type State struct {
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
}

type Store struct {
}

func NewStore() Store {
	return Store{}
}

func (Store) Set(dir string, s State) error {
	file, err := os.OpenFile(filepath.Join(dir, "state.json"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	err = encode(file, s)
	if err != nil {
		return err
	}

	return nil
}

func (Store) Get(dir string) (State, error) {
	s := State{}
	file, err := os.OpenFile(filepath.Join(dir, "state.json"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}

	err = json.NewDecoder(file).Decode(&s)
	if err != nil {
		return s, err
	}

	return s, nil
}

func encodeFile(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
