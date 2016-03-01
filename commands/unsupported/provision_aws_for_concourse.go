package unsupported

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type cloudformationSessionProvider interface {
	Session(aws.Config) (cloudformation.Session, error)
}

type ec2SessionProvider interface {
	Session(aws.Config) (ec2.Session, error)
}

type templateBuilder interface {
	Build() cloudformation.Template
}

type stackManager interface {
	CreateOrUpdate(cloudFormationClient cloudformation.Session, stackName string, template cloudformation.Template) error
	WaitForCompletion(cloudFormationClient cloudformation.Session, stackName string, sleepInterval time.Duration) error
}

type keyPairManager interface {
	Sync(ec2Session ec2.Session, keypair ec2.KeyPair) (ec2.KeyPair, error)
}

type ProvisionAWSForConcourse struct {
	builder                       templateBuilder
	stackManager                  stackManager
	keyPairManager                keyPairManager
	cloudformationSessionProvider cloudformationSessionProvider
	ec2SessionProvider            ec2SessionProvider
}

func NewProvisionAWSForConcourse(builder templateBuilder, stackManager stackManager, keyPairManager keyPairManager, cloudformationSessionProvider cloudformationSessionProvider, ec2SessionProvider ec2SessionProvider) ProvisionAWSForConcourse {
	return ProvisionAWSForConcourse{
		builder:                       builder,
		stackManager:                  stackManager,
		keyPairManager:                keyPairManager,
		cloudformationSessionProvider: cloudformationSessionProvider,
		ec2SessionProvider:            ec2SessionProvider,
	}
}

func (p ProvisionAWSForConcourse) Execute(globalFlags commands.GlobalFlags, state storage.State) (storage.State, error) {
	if state.KeyPair == nil {
		state.KeyPair = &storage.KeyPair{}
	}

	keyPair := ec2.KeyPair{
		Name:       state.KeyPair.Name,
		PublicKey:  []byte(state.KeyPair.PublicKey),
		PrivateKey: []byte(state.KeyPair.PrivateKey),
	}

	ec2Session, err := p.ec2SessionProvider.Session(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	keyPair, err = p.keyPairManager.Sync(ec2Session, keyPair)
	if err != nil {
		return state, err
	}

	state.KeyPair = &storage.KeyPair{
		Name:       keyPair.Name,
		PublicKey:  string(keyPair.PublicKey),
		PrivateKey: string(keyPair.PrivateKey),
	}

	template := p.builder.Build()
	template.SetKeyPairName(state.KeyPair.Name)

	cloudFormationSession, err := p.cloudformationSessionProvider.Session(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	if err := p.stackManager.CreateOrUpdate(cloudFormationSession, "concourse", template); err != nil {
		return state, err
	}

	if err := p.stackManager.WaitForCompletion(cloudFormationSession, "concourse", 2*time.Second); err != nil {
		return state, err
	}

	return state, nil
}
