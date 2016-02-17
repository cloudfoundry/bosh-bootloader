package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Store struct {
}

func NewStore() Store {
	return Store{}
}

func (s Store) Merge(dir string, newStateMap map[string]interface{}) error {
	path := filepath.Join(dir, "state.json")
	stateMap, err := s.read(path)
	if err != nil {
		return err
	}

	for k, v := range newStateMap {
		stateMap[k] = v
	}

	err = s.write(path, stateMap)
	if err != nil {
		return err
	}

	return nil
}

func (s Store) read(file string) (map[string]interface{}, error) {
	stateMap := make(map[string]interface{})

	_, err := os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return stateMap, err
		}
	} else {
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			return stateMap, err
		}

		err = json.Unmarshal(contents, &stateMap)
		if err != nil {
			return stateMap, err
		}
	}

	return stateMap, nil
}

func (s Store) write(file string, stateMap map[string]interface{}) error {
	data, err := json.Marshal(stateMap)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (s Store) GetString(dir string, key string) (string, bool, error) {
	stateMap, err := s.read(filepath.Join(dir, "state.json"))
	if err != nil {
		return "", false, err
	}

	v, ok := stateMap[key]
	if !ok {
		return "", false, nil
	}

	value, ok := v.(string)
	if !ok {
		return "", false, fmt.Errorf("value at key %q is not type %q", key, "string")
	}

	return value, true, nil
}
