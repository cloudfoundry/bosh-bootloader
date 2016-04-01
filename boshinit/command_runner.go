package boshinit

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const OS_READ_WRITE_MODE = os.FileMode(0644)

type CommandRunner struct {
	directory string
	command   executable
}

type State map[string]interface{}

type executable interface {
	Run() error
}

func NewCommandRunner(dir string, command executable) CommandRunner {
	return CommandRunner{
		directory: dir,
		command:   command,
	}
}

func (r CommandRunner) Execute(manifest []byte, privateKey string, state State) (State, error) {
	stateJSONPath := filepath.Join(r.directory, "bosh-state.json")

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return State{}, err
	}

	err = ioutil.WriteFile(stateJSONPath, stateJSON, OS_READ_WRITE_MODE)
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

	err = r.command.Run()
	if err != nil {
		return State{}, err
	}

	_, err = os.Stat(stateJSONPath)
	if err != nil {
		return State{}, nil
	}

	boshStateData, err := ioutil.ReadFile(stateJSONPath)
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
