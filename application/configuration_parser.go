package application

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type commandLineParser interface {
	Parse(arguments []string) (CommandLineConfiguration, error)
}

type awsCredentialValidator interface {
	Validate(accessKeyID string, secretAccessKey string, region string) error
}

type stateStore interface {
	Get(stateDirectory string) (storage.State, error)
	Set(stateDirectory string, state storage.State) error
}

type ConfigurationParser struct {
	commandLineParser      commandLineParser
	awsCredentialValidator awsCredentialValidator
	stateStore             stateStore
}

func NewConfigurationParser(commandLineParser commandLineParser, awsCredentialValidator awsCredentialValidator, stateStore stateStore) ConfigurationParser {
	return ConfigurationParser{
		commandLineParser:      commandLineParser,
		awsCredentialValidator: awsCredentialValidator,
		stateStore:             stateStore,
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

	err = p.validateCredentialsIfRequired(configuration.Command, configuration.State.AWS)
	if err != nil {
		return Configuration{}, err
	}

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

func (p ConfigurationParser) validateCredentialsIfRequired(command string, awsState storage.AWS) error {
	credentialsRequired := map[string]bool{
		"unsupported-deploy-bosh-on-aws-for-concourse": true,
		"destroy":                true,
		"unsupported-create-lbs": true,
		"unsupported-update-lbs": true,

		"help":              false,
		"version":           false,
		"director-address":  false,
		"director-username": false,
		"director-password": false,
		"ssh-key":           false,
	}

	required, commandFound := credentialsRequired[command]
	if !commandFound {
		return p.commandError(command)
	}

	if required {
		err := p.awsCredentialValidator.Validate(
			awsState.AccessKeyID,
			awsState.SecretAccessKey,
			awsState.Region,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p ConfigurationParser) commandError(command string) error {
	if command == "" {
		command = "[EMPTY]"
	}
	return NewInvalidCommandError(fmt.Errorf("unknown command: %s", command))
}
