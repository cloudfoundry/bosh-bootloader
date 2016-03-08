package unsupported

import (
	"io"
	"time"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type awsClientProvider interface {
	CloudFormationClient(aws.Config) (cloudformation.Client, error)
	EC2Client(aws.Config) (ec2.Client, error)
}

type templateBuilder interface {
	Build(keypairName string) templates.Template
}

type stackManager interface {
	CreateOrUpdate(cloudFormationClient cloudformation.Client, stackName string, template templates.Template) error
	WaitForCompletion(cloudFormationClient cloudformation.Client, stackName string, sleepInterval time.Duration) error
}

type keyPairManager interface {
	Sync(ec2Client ec2.Client, keypair ec2.KeyPair) (ec2.KeyPair, error)
}

type boshInitManifestBuilder interface {
	Build() boshinit.Manifest
}

type DeployBOSHOnAWSForConcourse struct {
	builder                 templateBuilder
	stackManager            stackManager
	keyPairManager          keyPairManager
	awsClientProvider       awsClientProvider
	boshInitManifestBuilder boshInitManifestBuilder
	stdout                  io.Writer
}

func NewDeployBOSHOnAWSForConcourse(builder templateBuilder, stackManager stackManager, keyPairManager keyPairManager, awsClientProvider awsClientProvider, boshInitManifestBuilder boshInitManifestBuilder, stdout io.Writer) DeployBOSHOnAWSForConcourse {
	return DeployBOSHOnAWSForConcourse{
		builder:                 builder,
		stackManager:            stackManager,
		keyPairManager:          keyPairManager,
		awsClientProvider:       awsClientProvider,
		boshInitManifestBuilder: boshInitManifestBuilder,
		stdout:                  stdout,
	}
}

func (d DeployBOSHOnAWSForConcourse) Execute(globalFlags commands.GlobalFlags, state storage.State) (storage.State, error) {
	if state.KeyPair == nil {
		state.KeyPair = &storage.KeyPair{}
	}

	keyPair := ec2.KeyPair{
		Name:       state.KeyPair.Name,
		PublicKey:  []byte(state.KeyPair.PublicKey),
		PrivateKey: []byte(state.KeyPair.PrivateKey),
	}

	ec2Client, err := d.awsClientProvider.EC2Client(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	keyPair, err = d.keyPairManager.Sync(ec2Client, keyPair)
	if err != nil {
		return state, err
	}

	state.KeyPair = &storage.KeyPair{
		Name:       keyPair.Name,
		PublicKey:  string(keyPair.PublicKey),
		PrivateKey: string(keyPair.PrivateKey),
	}

	template := d.builder.Build(state.KeyPair.Name)

	cloudFormationClient, err := d.awsClientProvider.CloudFormationClient(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	if err := d.stackManager.CreateOrUpdate(cloudFormationClient, "concourse", template); err != nil {
		return state, err
	}

	if err := d.stackManager.WaitForCompletion(cloudFormationClient, "concourse", 15*time.Second); err != nil {
		return state, err
	}

	manifest := d.boshInitManifestBuilder.Build()
	yaml, err := candiedyaml.Marshal(manifest)
	if err != nil {
		return state, err
	}
	d.stdout.Write([]byte("\nbosh-init manifest:\n\n"))
	d.stdout.Write(yaml)

	return state, nil
}
