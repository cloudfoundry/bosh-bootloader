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
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
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
	Describe(cloudFormationClient cloudformation.Client, name string) (cloudformation.Stack, error)
}

type keyPairManager interface {
	Sync(ec2Client ec2.Client, keypair ec2.KeyPair) (ec2.KeyPair, error)
}

type boshInitManifestBuilder interface {
	Build(boshinit.ManifestProperties) (boshinit.Manifest, error)
}

type DeployBOSHOnAWSForConcourse struct {
	templateBuilder         templateBuilder
	stackManager            stackManager
	keyPairManager          keyPairManager
	awsClientProvider       awsClientProvider
	boshInitManifestBuilder boshInitManifestBuilder
	stdout                  io.Writer
}

func NewDeployBOSHOnAWSForConcourse(templateBuilder templateBuilder, stackManager stackManager, keyPairManager keyPairManager, awsClientProvider awsClientProvider, boshInitManifestBuilder boshInitManifestBuilder, stdout io.Writer) DeployBOSHOnAWSForConcourse {
	return DeployBOSHOnAWSForConcourse{
		templateBuilder:         templateBuilder,
		stackManager:            stackManager,
		keyPairManager:          keyPairManager,
		awsClientProvider:       awsClientProvider,
		boshInitManifestBuilder: boshInitManifestBuilder,
		stdout:                  stdout,
	}
}

func (d DeployBOSHOnAWSForConcourse) Execute(globalFlags commands.GlobalFlags, state storage.State) (storage.State, error) {
	cloudFormationClient, err := d.awsClientProvider.CloudFormationClient(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	state, err = d.synchronizeKeyPairs(state, globalFlags.EndpointOverride)
	if err != nil {
		return state, err
	}

	err = d.createInfrastructure(state.KeyPair.Name, cloudFormationClient)
	if err != nil {
		return state, err
	}

	directorSSLKeyPair := ssl.KeyPair{}
	if state.BOSH != nil {
		directorSSLKeyPair.Certificate = []byte(state.BOSH.DirectorSSLCertificate)
		directorSSLKeyPair.PrivateKey = []byte(state.BOSH.DirectorSSLPrivateKey)
	}

	directorSSLKeyPair, err = d.generateBoshInitManifest(cloudFormationClient, state.AWS.Region, state.KeyPair.Name, directorSSLKeyPair)
	if err != nil {
		return state, err
	}

	if state.BOSH == nil {
		state.BOSH = &storage.BOSH{
			DirectorSSLCertificate: string(directorSSLKeyPair.Certificate),
			DirectorSSLPrivateKey:  string(directorSSLKeyPair.PrivateKey),
		}
	}

	return state, nil
}

func (d DeployBOSHOnAWSForConcourse) synchronizeKeyPairs(state storage.State, endpointOverride string) (storage.State, error) {
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
		EndpointOverride: endpointOverride,
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

	return state, nil
}

func (d DeployBOSHOnAWSForConcourse) createInfrastructure(keyPairName string, cloudFormationClient cloudformation.Client) error {
	template := d.templateBuilder.Build(keyPairName)

	if err := d.stackManager.CreateOrUpdate(cloudFormationClient, "concourse", template); err != nil {
		return err
	}

	if err := d.stackManager.WaitForCompletion(cloudFormationClient, "concourse", 15*time.Second); err != nil {
		return err
	}

	return nil
}

func (d DeployBOSHOnAWSForConcourse) generateBoshInitManifest(cloudFormationClient cloudformation.Client, region string, keyPairName string, keyPair ssl.KeyPair) (ssl.KeyPair, error) {
	stack, err := d.stackManager.Describe(cloudFormationClient, "concourse")
	if err != nil {
		return ssl.KeyPair{}, err
	}

	manifestProperties := boshinit.ManifestProperties{
		SubnetID:         stack.Outputs["BOSHSubnet"],
		AvailabilityZone: stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        stack.Outputs["BOSHEIP"],
		AccessKeyID:      stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    stack.Outputs["BOSHSecurityGroup"],
		Region:           region,
		DefaultKeyName:   keyPairName,
		SSLKeyPair:       keyPair,
	}

	manifest, err := d.boshInitManifestBuilder.Build(manifestProperties)
	if err != nil {
		return ssl.KeyPair{}, err
	}

	yaml, err := candiedyaml.Marshal(manifest)
	if err != nil {
		return ssl.KeyPair{}, err
	}

	d.stdout.Write([]byte("\nbosh-init manifest:\n\n"))
	d.stdout.Write(yaml)

	return manifest.DirectorSSLKeyPair(), nil
}
