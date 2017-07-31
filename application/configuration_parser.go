package application

import "github.com/cloudfoundry/bosh-bootloader/storage"

var getState func(string) (storage.State, error) = storage.GetState

type commandLineParser interface {
	Parse(arguments []string) (CommandLineConfiguration, error)
}

type stateStore interface {
	Set(state storage.State) error
}

type ConfigurationParser struct {
	commandLineParser commandLineParser
}

func NewConfigurationParser(commandLineParser commandLineParser) ConfigurationParser {
	return ConfigurationParser{
		commandLineParser: commandLineParser,
	}
}

func (p ConfigurationParser) Parse(arguments []string) (Configuration, error) {
	commandLineConfiguration, err := p.commandLineParser.Parse(arguments)
	if err != nil {
		return Configuration{}, err
	}

	configuration := Configuration{
		Global: GlobalConfiguration{
			StateDir: commandLineConfiguration.StateDir,
			Debug:    commandLineConfiguration.Debug,
		},
		Command:         commandLineConfiguration.Command,
		SubcommandFlags: commandLineConfiguration.SubcommandFlags,
		State:           storage.State{},
	}

	if !p.isHelpOrVersion(configuration.Command, configuration.SubcommandFlags) {
		configuration.State, err = getState(configuration.Global.StateDir)
		if err != nil {
			return Configuration{}, err
		}
	}

	return configuration, nil
}

func (ConfigurationParser) isHelpOrVersion(command string, subcommandFlags StringSlice) bool {
	if command == "help" || command == "version" {
		return true
	}

	if subcommandFlags.ContainsAny("--help", "-h", "--version", "-v") {
		return true
	}

	return false
}
