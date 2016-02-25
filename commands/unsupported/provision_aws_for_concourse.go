package unsupported

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type cloudformationSessionProvider interface {
	Session(aws.Config) (cloudformation.Session, error)
}

type cloudformationCreator interface {
	CreateOrUpdate(cloudFormationClient cloudformation.Session, stackName string, template cloudformation.Template) error
}

type ProvisionAWSForConcourse struct {
	builder  templateBuilder
	manager  cloudformationCreator
	provider cloudformationSessionProvider
}

func NewProvisionAWSForConcourse(builder templateBuilder, manager cloudformationCreator, provider cloudformationSessionProvider) ProvisionAWSForConcourse {
	return ProvisionAWSForConcourse{
		builder:  builder,
		manager:  manager,
		provider: provider,
	}
}

func (p ProvisionAWSForConcourse) Execute(globalFlags commands.GlobalFlags, state storage.State) (storage.State, error) {
	template := p.builder.Build()

	if state.KeyPair == nil {
		return state, errors.New("no keypair is present, you can generate a keypair by running the unsupported-create-bosh-aws-keypair command.")
	}

	template.SetKeyPairName(state.KeyPair.Name)

	session, err := p.provider.Session(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	if err := p.manager.CreateOrUpdate(session, "concourse", template); err != nil {
		return state, err
	}

	return state, nil
}
