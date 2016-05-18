package application

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type CommandSet map[string]commands.Command

type App struct {
	commands      CommandSet
	configuration Configuration
	stateStore    stateStore
	usage         func()
}

func New(commands CommandSet, configuration Configuration, stateStore stateStore, usage func()) App {
	return App{
		commands:      commands,
		configuration: configuration,
		stateStore:    stateStore,
		usage:         usage,
	}
}

func (a App) Run() error {
	newState, err := a.execute(a.configuration)
	if err != nil {
		return err
	}

	err = a.stateStore.Set(a.configuration.Global.StateDir, newState)
	if err != nil {
		return err
	}

	return nil
}

func (a App) execute(configuration Configuration) (storage.State, error) {
	command, ok := a.commands[configuration.Command]
	if !ok {
		return storage.State{}, errors.New("command not found")
	}

	globalFlags := commands.GlobalFlags{
		StateDir:           configuration.Global.StateDir,
		EndpointOverride:   configuration.Global.EndpointOverride,
		AWSAccessKeyID:     configuration.State.AWS.AccessKeyID,
		AWSSecretAccessKey: configuration.State.AWS.SecretAccessKey,
		AWSRegion:          configuration.State.AWS.Region,
	}

	state, err := command.Execute(globalFlags, configuration.SubcommandFlags, configuration.State)
	if err != nil {
		return storage.State{}, err
	}

	return state, nil
}
