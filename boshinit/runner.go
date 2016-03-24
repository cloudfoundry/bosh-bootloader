package boshinit

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const OS_READ_WRITE_MODE = os.FileMode(0644)

type Runner struct {
	directory string
	command   executable
	logger    logger
}

type State map[string]interface{}

type executable interface {
	Run() error
}

func NewRunner(dir string, command executable, logger logger) Runner {
	return Runner{
		directory: dir,
		command:   command,
		logger:    logger,
	}
}

func (r Runner) Deploy(manifest []byte, privateKey string, state State) (State, error) {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return State{}, err
	}

	err = ioutil.WriteFile(filepath.Join(r.directory, "bosh-state.json"), stateJSON, OS_READ_WRITE_MODE)
	if err != nil {
		return State{}, err
	}

	err = ioutil.WriteFile(filepath.Join(r.directory, "bosh.yml"), manifest, OS_READ_WRITE_MODE)
	if err != nil {
		return State{}, err
	}

	err = ioutil.WriteFile(filepath.Join(r.directory, "bosh.pem"), []byte(privateKey), OS_READ_WRITE_MODE)
	if err != nil {
		return State{}, err
	}

	r.logger.Step("deploying bosh director")
	err = r.command.Run()
	if err != nil {
		return State{}, err
	}

	boshStateData, err := ioutil.ReadFile(filepath.Join(r.directory, "bosh-state.json"))
	if err != nil {
		return State{}, err
	}

	state = State{}
	err = json.Unmarshal(boshStateData, &state)
	if err != nil {
		return State{}, err
	}

	return state, nil
}
