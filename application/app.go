package application

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
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
	err := a.execute()
	if err != nil {
		return err
	}

	return nil
}

func (a App) execute() error {
	command, ok := a.commands[a.configuration.Command]
	if !ok {
		a.usage()
		return fmt.Errorf("unknown command: %s", a.configuration.Command)
	}

	err := command.Execute(a.configuration.SubcommandFlags, a.configuration.State)
	if err != nil {
		return err
	}

	return nil
}
