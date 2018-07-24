package application

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type CommandSet map[string]commands.Command

type usage interface {
	Print()
	PrintCommandUsage(command, message string)
}

type App struct {
	commands      CommandSet
	configuration Configuration
	usage         usage
}

func New(commands CommandSet, configuration Configuration, usage usage) App {
	return App{
		commands:      commands,
		configuration: configuration,
		usage:         usage,
	}
}

func (a App) Run() error {
	err := a.execute()
	if _, ok := err.(commands.ExitSuccessfully); !ok {
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

	if a.configuration.ShowCommandHelp {
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

	if a.configuration.Command == "--version" || a.configuration.Command == "-v" || a.configuration.SubcommandFlags.ContainsAny("--version", "-v") {
		versionCommand, err := a.getCommand("version")
		if err != nil {
			return err
		}

		return versionCommand.Execute([]string{}, storage.State{})
	}

	if (a.configuration.Command == "plan" || a.configuration.Command == "up") && a.configuration.Global.Name != "" {
		a.configuration.SubcommandFlags = append(a.configuration.SubcommandFlags, "--name", a.configuration.Global.Name)
	}

	err = command.CheckFastFails(a.configuration.SubcommandFlags, a.configuration.State)
	if err != nil {
		return err
	}

	return command.Execute(a.configuration.SubcommandFlags, a.configuration.State)
}
