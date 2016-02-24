package unsupported

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"
)

type cloudformationSessionProvider interface {
	Session(aws.Config) (cloudformation.Session, error)
}

type cloudformationCreator interface {
	Create(cloudFormationClient cloudformation.Session, stackName string, template cloudformation.Template) error
}

type ProvisionAWSForConcourse struct {
	stateStore stateStore
	builder    templateBuilder
	creator    cloudformationCreator
	provider   cloudformationSessionProvider
}

func NewProvisionAWSForConcourse(stateStore stateStore, builder templateBuilder, creator cloudformationCreator, provider cloudformationSessionProvider) ProvisionAWSForConcourse {
	return ProvisionAWSForConcourse{
		stateStore: stateStore,
		builder:    builder,
		creator:    creator,
		provider:   provider,
	}
}

func (p ProvisionAWSForConcourse) Execute(globalFlags commands.GlobalFlags) error {
	state, err := p.stateStore.Get(globalFlags.StateDir)
	if err != nil {
		return err
	}

	config := getAWSConfig(state, aws.Config{
		AccessKeyID:      globalFlags.AWSAccessKeyID,
		SecretAccessKey:  globalFlags.AWSSecretAccessKey,
		Region:           globalFlags.AWSRegion,
		EndpointOverride: globalFlags.EndpointOverride,
	})

	template := p.builder.Build()

	if state.KeyPair == nil {
		return errors.New("no keypair is present, you can generate a keypair by running the unsupported-create-bosh-aws-keypair command.")
	}

	template.SetKeyPairName(state.KeyPair.Name)

	session, err := p.provider.Session(config)
	if err != nil {
		return err
	}

	if err := p.creator.Create(session, "concourse", template); err != nil {
		return err
	}

	return nil
}

func getAWSConfig(state state.State, config aws.Config) aws.Config {
	if config.AccessKeyID == "" {
		config.AccessKeyID = state.AWS.AccessKeyID
	}

	if config.SecretAccessKey == "" {
		config.SecretAccessKey = state.AWS.SecretAccessKey
	}

	if config.Region == "" {
		config.Region = state.AWS.Region
	}

	return config
}
