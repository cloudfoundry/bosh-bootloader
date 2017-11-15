package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	if state.TFState == "" {
		return state, nil
	}

	varsDir, err := m.store.GetVarsDir()
	if err != nil {
		return State{}, fmt.Errorf("migrating terraform state: %s", err)
	}
	err = ioutil.WriteFile(filepath.Join(varsDir, "terraform.tfstate"), []byte(state.TFState), os.ModePerm)
	if err != nil {
		return State{}, fmt.Errorf("migrating terraform state: %s", err)
	}
	state.TFState = ""

	err = m.store.Set(state)
	if err != nil {
		return State{}, fmt.Errorf("saving migrated state: %s", err)
	}

	return state, nil
}
