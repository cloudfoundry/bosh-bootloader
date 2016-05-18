package application

import (
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

var getwd func() (string, error) = os.Getwd

type stateStore interface {
	Get(stateDirectory string) (storage.State, error)
	Set(stateDirectory string, state storage.State) error
}

type ConfigurationParser struct {
	stateStore stateStore
}

func NewConfigurationParser(stateStore stateStore) ConfigurationParser {
	return ConfigurationParser{
		stateStore: stateStore,
	}
}

func (p ConfigurationParser) Parse(arguments []string) (Configuration, error) {
	configuration := Configuration{
		Global:          GlobalConfiguration{},
		Command:         "",
		SubcommandFlags: []string{},
	}

	configuration, err := p.parseCommandLineArguments(configuration, arguments)
	if err != nil {
		return Configuration{}, err
	}

	configuration.State, err = p.stateStore.Get(configuration.Global.StateDir)
	if err != nil {
		return Configuration{}, err
	}

	configuration.State.AWS = p.overrideAWSCredentials(configuration.Global, configuration.State.AWS)

	return configuration, nil
}

func (p ConfigurationParser) parseCommandLineArguments(configuration Configuration, arguments []string) (Configuration, error) {
	var err error
	var remainingArguments []string
	configuration.Global, remainingArguments, err = p.parseGlobalFlags(configuration.Global, arguments)
	if err != nil {
		return Configuration{}, err
	}

	configuration = p.convertFlagsToCommands(configuration)

	if configuration.Command == "" {
		configuration, err = p.parseCommandAndSubcommandFlags(configuration, remainingArguments)
		if err != nil {
			return Configuration{}, err
		}
	}

	configuration, err = p.setDefaultStateDirectory(configuration)
	if err != nil {
		return Configuration{}, err
	}

	return configuration, nil
}

func (ConfigurationParser) parseGlobalFlags(globalConfiguration GlobalConfiguration, arguments []string) (GlobalConfiguration, []string, error) {
	globalFlags := flags.New("global")

	globalFlags.Bool(&globalConfiguration.Help, "h", "help", false)
	globalFlags.Bool(&globalConfiguration.Version, "v", "version", false)
	globalFlags.String(&globalConfiguration.EndpointOverride, "endpoint-override", "")
	globalFlags.String(&globalConfiguration.AWSAccessKeyID, "aws-access-key-id", "")
	globalFlags.String(&globalConfiguration.AWSSecretAccessKey, "aws-secret-access-key", "")
	globalFlags.String(&globalConfiguration.AWSRegion, "aws-region", "")
	globalFlags.String(&globalConfiguration.StateDir, "state-dir", "")

	err := globalFlags.Parse(arguments)
	if err != nil {
		return GlobalConfiguration{}, []string{}, NewInvalidFlagError(err)
	}

	return globalConfiguration, globalFlags.Args(), nil
}

func (ConfigurationParser) parseCommandAndSubcommandFlags(configuration Configuration, remainingArguments []string) (Configuration, error) {
	if len(remainingArguments) == 0 {
		return Configuration{}, NewCommandNotProvidedError()
	}

	configuration.Command = remainingArguments[0]
	configuration.SubcommandFlags = remainingArguments[1:]

	return configuration, nil
}

func (ConfigurationParser) setDefaultStateDirectory(configuration Configuration) (Configuration, error) {
	if configuration.Global.StateDir == "" {
		wd, err := getwd()
		if err != nil {
			return Configuration{}, err
		}

		configuration.Global.StateDir = wd
	}

	return configuration, nil
}

func (ConfigurationParser) convertFlagsToCommands(configuration Configuration) Configuration {
	if configuration.Global.Version {
		configuration.Command = "version"
	}

	if configuration.Global.Help {
		configuration.Command = "help"
	}

	return configuration
}

func (ConfigurationParser) overrideAWSCredentials(globalConfiguration GlobalConfiguration, awsState storage.AWS) storage.AWS {
	if globalConfiguration.AWSAccessKeyID != "" {
		awsState.AccessKeyID = globalConfiguration.AWSAccessKeyID
	}

	if globalConfiguration.AWSSecretAccessKey != "" {
		awsState.SecretAccessKey = globalConfiguration.AWSSecretAccessKey
	}

	if globalConfiguration.AWSRegion != "" {
		awsState.Region = globalConfiguration.AWSRegion
	}

	return awsState
}
