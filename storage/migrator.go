package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
)

type store interface {
	Set(state State) error
	GetVarsDir() (string, error)
}

type Migrator struct {
	store store
}

func NewMigrator(store store) Migrator {
	return Migrator{store: store}
}

func (m Migrator) Migrate(state State) (State, error) {
	if reflect.DeepEqual(state, State{}) {
		return state, nil
	}

	varsDir, err := m.store.GetVarsDir()
	if err != nil {
		return State{}, fmt.Errorf("migrating state: %s", err)
	}
	if state.TFState != "" {
		err = ioutil.WriteFile(filepath.Join(varsDir, "terraform.tfstate"), []byte(state.TFState), os.ModePerm)
		if err != nil {
			return State{}, fmt.Errorf("migrating terraform state: %s", err)
		}
		state.TFState = ""
	}

	if len(state.BOSH.State) > 0 {
		stateJSON, err := json.Marshal(state.BOSH.State)
		if err != nil {
			return State{}, fmt.Errorf("marshalling bosh state: %s", err)
		}
		err = ioutil.WriteFile(filepath.Join(varsDir, "bosh-state.json"), stateJSON, os.ModePerm)
		if err != nil {
			return State{}, fmt.Errorf("migrating bosh state: %s", err)
		}
		state.BOSH.State = nil
	}

	err = m.store.Set(state)
	if err != nil {
		return State{}, fmt.Errorf("saving migrated state: %s", err)
	}

	return state, nil
}
