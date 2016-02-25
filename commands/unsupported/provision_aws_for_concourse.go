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
	builder  templateBuilder
	creator  cloudformationCreator
	provider cloudformationSessionProvider
}

func NewProvisionAWSForConcourse(builder templateBuilder, creator cloudformationCreator, provider cloudformationSessionProvider) ProvisionAWSForConcourse {
	return ProvisionAWSForConcourse{
		builder:  builder,
		creator:  creator,
		provider: provider,
	}
}

func (p ProvisionAWSForConcourse) Execute(globalFlags commands.GlobalFlags, s state.State) (state.State, error) {
	template := p.builder.Build()

	if s.KeyPair == nil {
		return s, errors.New("no keypair is present, you can generate a keypair by running the unsupported-create-bosh-aws-keypair command.")
	}

	template.SetKeyPairName(s.KeyPair.Name)

	session, err := p.provider.Session(aws.Config{
		AccessKeyID:      s.AWS.AccessKeyID,
		SecretAccessKey:  s.AWS.SecretAccessKey,
		Region:           s.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return s, err
	}

	if err := p.creator.Create(session, "concourse", template); err != nil {
		return s, err
	}

	return s, nil
}
