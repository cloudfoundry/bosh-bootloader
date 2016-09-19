package application

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
)

type CommandSet map[string]commands.Command

type usage interface {
	Print()
	PrintCommandUsage(command, message string)
}

type App struct {
	commands      CommandSet
	configuration Configuration
	stateStore    stateStore
	usage         usage
}

func New(commands CommandSet, configuration Configuration, stateStore stateStore,
	usage usage) App {
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

func (a App) getCommand(commandString string) (commands.Command, error) {
	command, ok := a.commands[commandString]
	if !ok {
		a.usage.Print()
		return nil, fmt.Errorf("unknown command: %s", commandString)
	}
	return command, nil
}

func (a App) execute() error {
	command, err := a.getCommand(a.configuration.Command)
	if err != nil {
		return err
	}

	if a.configuration.SubcommandFlags.ContainsAny("--help", "-h") {
		a.usage.PrintCommandUsage(a.configuration.Command, command.Usage())
		return nil
	}

	if a.configuration.Command == "help" && len(a.configuration.SubcommandFlags) != 0 {
		commandString := a.configuration.SubcommandFlags[0]
		command, err = a.getCommand(commandString)
		if err != nil {
			return err
		}
		a.usage.PrintCommandUsage(commandString, command.Usage())
		return nil
	}

	err = command.Execute(a.configuration.SubcommandFlags, a.configuration.State)
	if err != nil {
		return err
	}

	return nil
}
