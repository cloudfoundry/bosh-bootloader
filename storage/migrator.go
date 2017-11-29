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
		err = ioutil.WriteFile(filepath.Join(varsDir, "terraform.tfstate"), []byte(state.TFState), StateMode)
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
		err = ioutil.WriteFile(filepath.Join(varsDir, "bosh-state.json"), stateJSON, StateMode)
		if err != nil {
			return State{}, fmt.Errorf("migrating bosh state: %s", err)
		}
		state.BOSH.State = nil
	}

	if len(state.Jumpbox.State) > 0 {
		stateJSON, err := json.Marshal(state.Jumpbox.State)
		if err != nil {
			return State{}, fmt.Errorf("marshalling jumpbox state: %s", err)
		}
		err = ioutil.WriteFile(filepath.Join(varsDir, "jumpbox-state.json"), stateJSON, StateMode)
		if err != nil {
			return State{}, fmt.Errorf("migrating jumpbox state: %s", err)
		}
		state.Jumpbox.State = nil
	}

	legacyDirectorVarsStore := filepath.Join(varsDir, "director-variables.yml")
	if _, err := os.Stat(legacyDirectorVarsStore); err == nil {
		boshVars, err := ioutil.ReadFile(legacyDirectorVarsStore)
		if err != nil {
			return State{}, fmt.Errorf("reading legacy director vars store: %s", err)
		}

		state.BOSH.Variables = string(boshVars)

		if err := os.Remove(legacyDirectorVarsStore); err != nil {
			return State{}, fmt.Errorf("removing legacy director vars store: %s", err) //not tested
		}
	}

	if state.BOSH.Variables != "" {
		err = ioutil.WriteFile(filepath.Join(varsDir, "director-vars-store.yml"), []byte(state.BOSH.Variables), StateMode)
		if err != nil {
			return State{}, fmt.Errorf("migrating bosh variables: %s", err)
		}
		state.BOSH.Variables = ""
	}

	legacyJumpboxVarsStore := filepath.Join(varsDir, "jumpbox-variables.yml")
	if _, err := os.Stat(legacyJumpboxVarsStore); err == nil {
		jumpboxVars, err := ioutil.ReadFile(legacyJumpboxVarsStore)
		if err != nil {
			return State{}, fmt.Errorf("reading legacy jumpbox vars store: %s", err)
		}

		state.Jumpbox.Variables = string(jumpboxVars)

		if err := os.Remove(legacyJumpboxVarsStore); err != nil {
			return State{}, fmt.Errorf("removing legacy jumpbox vars store: %s", err) //not tested
		}
	}

	if state.Jumpbox.Variables != "" {
		err = ioutil.WriteFile(filepath.Join(varsDir, "jumpbox-vars-store.yml"), []byte(state.Jumpbox.Variables), StateMode)
		if err != nil {
			return State{}, fmt.Errorf("migrating jumpbox variables: %s", err)
		}
		state.Jumpbox.Variables = ""
	}

	err = m.store.Set(state)
	if err != nil {
		return State{}, fmt.Errorf("saving migrated state: %s", err)
	}

	return state, nil
}
