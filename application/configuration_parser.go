package application

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type commandLineParser interface {
	Parse(arguments []string) (CommandLineConfiguration, error)
}

type stateStore interface {
	Get(stateDirectory string) (storage.State, error)
	Set(stateDirectory string, state storage.State) error
}

type ConfigurationParser struct {
	commandLineParser commandLineParser
	stateStore        stateStore
}

func NewConfigurationParser(commandLineParser commandLineParser, stateStore stateStore) ConfigurationParser {
	return ConfigurationParser{
		commandLineParser: commandLineParser,
		stateStore:        stateStore,
	}
}

func (p ConfigurationParser) Parse(arguments []string) (Configuration, error) {
	commandLineConfiguration, err := p.commandLineParser.Parse(arguments)
	if err != nil {
		return Configuration{}, err
	}

	configuration := Configuration{
		Global: GlobalConfiguration{
			StateDir:         commandLineConfiguration.StateDir,
			EndpointOverride: commandLineConfiguration.EndpointOverride,
		},
		Command:         commandLineConfiguration.Command,
		SubcommandFlags: commandLineConfiguration.SubcommandFlags,
		State:           storage.State{},
	}

	configuration.State, err = p.stateStore.Get(configuration.Global.StateDir)
	if err != nil {
		return Configuration{}, err
	}

	configuration.State.AWS = p.overrideAWSCredentials(commandLineConfiguration, configuration.State.AWS)

	return configuration, nil
}

func (ConfigurationParser) overrideAWSCredentials(commandLineConfiguration CommandLineConfiguration, awsState storage.AWS) storage.AWS {
	if commandLineConfiguration.AWSAccessKeyID != "" {
		awsState.AccessKeyID = commandLineConfiguration.AWSAccessKeyID
	}

	if commandLineConfiguration.AWSSecretAccessKey != "" {
		awsState.SecretAccessKey = commandLineConfiguration.AWSSecretAccessKey
	}

	if commandLineConfiguration.AWSRegion != "" {
		awsState.Region = commandLineConfiguration.AWSRegion
	}

	return awsState
}
