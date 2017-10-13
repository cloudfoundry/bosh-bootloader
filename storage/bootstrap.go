package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

type logger interface {
	Println(message string)
}

type StateBootstrap struct {
	logger logger
}

func NewStateBootstrap(logger logger) StateBootstrap {
	return StateBootstrap{
		logger: logger,
	}
}

func (b StateBootstrap) GetState(dir string) (State, error) {
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
		b.logger.Println(fmt.Sprintf("Warning: Current schema version (%d) is newer than existing bbl environment schema (%d). Some things may not work as expected until you bbl up again.", STATE_VERSION, state.Version))
	}

	return state, nil
}
