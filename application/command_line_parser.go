package application

import (
	"errors"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/flags"
)

var getwd func() (string, error) = os.Getwd

type CommandLineConfiguration struct {
	Command          string
	SubcommandFlags  []string
	EndpointOverride string
	StateDir         string

	help    bool
	version bool
}

type CommandLineParser struct {
	usage func()
}

func NewCommandLineParser(usage func()) CommandLineParser {
	return CommandLineParser{
		usage: usage,
	}
}

func (p CommandLineParser) Parse(arguments []string) (CommandLineConfiguration, error) {
	var err error
	commandLineConfiguration := CommandLineConfiguration{}

	commandFinderResult := NewCommandFinder().FindCommand(arguments)

	commandLineConfiguration, _, err = p.parseGlobalFlags(commandLineConfiguration, commandFinderResult.GlobalFlags)
	if err != nil {
		return CommandLineConfiguration{}, err
	}

	commandLineConfiguration.Command = commandFinderResult.Command
	commandLineConfiguration = p.convertFlagsToCommands(commandLineConfiguration)
	if commandLineConfiguration.Command == "" {
		p.usage()
		return CommandLineConfiguration{}, errors.New("unknown command: [EMPTY]")
	}

	commandLineConfiguration.SubcommandFlags = commandFinderResult.OtherArgs
	commandLineConfiguration, err = p.setDefaultStateDirectory(commandLineConfiguration)
	if err != nil {
		return CommandLineConfiguration{}, err
	}

	return commandLineConfiguration, nil
}

func (c CommandLineParser) parseGlobalFlags(commandLineConfiguration CommandLineConfiguration, arguments []string) (CommandLineConfiguration, []string, error) {
	globalFlags := flags.New("global")

	globalFlags.String(&commandLineConfiguration.EndpointOverride, "endpoint-override", "")
	globalFlags.String(&commandLineConfiguration.StateDir, "state-dir", "")

	globalFlags.Bool(&commandLineConfiguration.help, "h", "help", false)
	globalFlags.Bool(&commandLineConfiguration.version, "v", "version", false)

	err := globalFlags.Parse(arguments)
	if err != nil {
		c.usage()
		return CommandLineConfiguration{}, []string{}, err
	}

	return commandLineConfiguration, globalFlags.Args(), nil
}

func (c CommandLineParser) parseCommandAndSubcommandFlags(commandLineConfiguration CommandLineConfiguration, remainingArguments []string) (CommandLineConfiguration, error) {
	if len(remainingArguments) == 0 {
		c.usage()
		return CommandLineConfiguration{}, errors.New("unknown command: [EMPTY]")
	}

	commandLineConfiguration.Command = remainingArguments[0]
	commandLineConfiguration.SubcommandFlags = remainingArguments[1:]

	return commandLineConfiguration, nil
}

func (CommandLineParser) setDefaultStateDirectory(commandLineConfiguration CommandLineConfiguration) (CommandLineConfiguration, error) {
	if commandLineConfiguration.StateDir == "" {
		wd, err := getwd()
		if err != nil {
			return CommandLineConfiguration{}, err
		}

		commandLineConfiguration.StateDir = wd
	}

	return commandLineConfiguration, nil
}

func (CommandLineParser) convertFlagsToCommands(commandLineConfiguration CommandLineConfiguration) CommandLineConfiguration {
	if commandLineConfiguration.version {
		commandLineConfiguration.Command = "version"
	}

	if commandLineConfiguration.help {
		commandLineConfiguration.Command = "help"
	}

	return commandLineConfiguration
}
