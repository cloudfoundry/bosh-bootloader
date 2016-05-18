package application

import (
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
)

var getwd func() (string, error) = os.Getwd

type CommandLineConfiguration struct {
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	Command            string
	SubcommandFlags    []string
	EndpointOverride   string
	StateDir           string

	help    bool
	version bool
}

type CommandLineParser struct {
}

func NewCommandLineParser() CommandLineParser {
	return CommandLineParser{}
}

func (p CommandLineParser) Parse(arguments []string) (CommandLineConfiguration, error) {
	var err error
	var remainingArguments []string
	commandLineConfiguration := CommandLineConfiguration{}

	commandLineConfiguration, remainingArguments, err = p.parseGlobalFlags(commandLineConfiguration, arguments)
	if err != nil {
		return CommandLineConfiguration{}, err
	}

	commandLineConfiguration = p.convertFlagsToCommands(commandLineConfiguration)

	if commandLineConfiguration.Command == "" {
		commandLineConfiguration, err = p.parseCommandAndSubcommandFlags(commandLineConfiguration, remainingArguments)
		if err != nil {
			return CommandLineConfiguration{}, err
		}
	}

	commandLineConfiguration, err = p.setDefaultStateDirectory(commandLineConfiguration)
	if err != nil {
		return CommandLineConfiguration{}, err
	}

	return commandLineConfiguration, nil
	if err != nil {
		return CommandLineConfiguration{}, err
	}

	return commandLineConfiguration, nil
}

func (CommandLineParser) parseGlobalFlags(commandLineConfiguration CommandLineConfiguration, arguments []string) (CommandLineConfiguration, []string, error) {
	globalFlags := flags.New("global")

	globalFlags.String(&commandLineConfiguration.EndpointOverride, "endpoint-override", "")
	globalFlags.String(&commandLineConfiguration.StateDir, "state-dir", "")

	globalFlags.Bool(&commandLineConfiguration.help, "h", "help", false)
	globalFlags.Bool(&commandLineConfiguration.version, "v", "version", false)
	globalFlags.String(&commandLineConfiguration.AWSAccessKeyID, "aws-access-key-id", "")
	globalFlags.String(&commandLineConfiguration.AWSSecretAccessKey, "aws-secret-access-key", "")
	globalFlags.String(&commandLineConfiguration.AWSRegion, "aws-region", "")

	err := globalFlags.Parse(arguments)
	if err != nil {
		return CommandLineConfiguration{}, []string{}, NewInvalidFlagError(err)
	}

	return commandLineConfiguration, globalFlags.Args(), nil
}

func (CommandLineParser) parseCommandAndSubcommandFlags(commandLineConfiguration CommandLineConfiguration, remainingArguments []string) (CommandLineConfiguration, error) {
	if len(remainingArguments) == 0 {
		return CommandLineConfiguration{}, NewCommandNotProvidedError()
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
