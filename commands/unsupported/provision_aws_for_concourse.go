package unsupported

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type cloudformationClientProvider interface {
	Client(aws.Config) (cloudformation.Client, error)
}

type ec2ClientProvider interface {
	Client(aws.Config) (ec2.Client, error)
}

type templateBuilder interface {
	Build(keypairName string) cloudformation.Template
}

type stackManager interface {
	CreateOrUpdate(cloudFormationClient cloudformation.Client, stackName string, template cloudformation.Template) error
	WaitForCompletion(cloudFormationClient cloudformation.Client, stackName string, sleepInterval time.Duration) error
}

type keyPairManager interface {
	Sync(ec2Client ec2.Client, keypair ec2.KeyPair) (ec2.KeyPair, error)
}

type ProvisionAWSForConcourse struct {
	builder                      templateBuilder
	stackManager                 stackManager
	keyPairManager               keyPairManager
	cloudformationClientProvider cloudformationClientProvider
	ec2ClientProvider            ec2ClientProvider
}

func NewProvisionAWSForConcourse(builder templateBuilder, stackManager stackManager, keyPairManager keyPairManager, cloudformationClientProvider cloudformationClientProvider, ec2ClientProvider ec2ClientProvider) ProvisionAWSForConcourse {
	return ProvisionAWSForConcourse{
		builder:                      builder,
		stackManager:                 stackManager,
		keyPairManager:               keyPairManager,
		cloudformationClientProvider: cloudformationClientProvider,
		ec2ClientProvider:            ec2ClientProvider,
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

	ec2Client, err := p.ec2ClientProvider.Client(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	keyPair, err = p.keyPairManager.Sync(ec2Client, keyPair)
	if err != nil {
		return state, err
	}

	state.KeyPair = &storage.KeyPair{
		Name:       keyPair.Name,
		PublicKey:  string(keyPair.PublicKey),
		PrivateKey: string(keyPair.PrivateKey),
	}

	template := p.builder.Build(state.KeyPair.Name)

	cloudFormationClient, err := p.cloudformationClientProvider.Client(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	if err := p.stackManager.CreateOrUpdate(cloudFormationClient, "concourse", template); err != nil {
		return state, err
	}

	if err := p.stackManager.WaitForCompletion(cloudFormationClient, "concourse", 2*time.Second); err != nil {
		return state, err
	}

	return state, nil
}
