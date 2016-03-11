package unsupported

import (
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type infrastructureCreator interface {
	Create(keyPairName string, client cloudformation.Client) error
}

type stackManager interface {
	CreateOrUpdate(cloudFormationClient cloudformation.Client, stackName string, template templates.Template) error
	WaitForCompletion(cloudFormationClient cloudformation.Client, stackName string, sleepInterval time.Duration) error
	Describe(cloudFormationClient cloudformation.Client, name string) (cloudformation.Stack, error)
}

type awsClientProvider interface {
	CloudFormationClient(aws.Config) (cloudformation.Client, error)
	EC2Client(aws.Config) (ec2.Client, error)
}

type keyPairSynchronizer interface {
	Sync(keypair KeyPair, ec2Client ec2.Client) (KeyPair, error)
}

type boshDeployer interface {
	Deploy(client cloudformation.Client, region, keyPairName string, directorSSLKeyPair ssl.KeyPair) (ssl.KeyPair, error)
}

type DeployBOSHOnAWSForConcourse struct {
	infrastructureCreator infrastructureCreator
	keyPairSynchronizer   keyPairSynchronizer
	awsClientProvider     awsClientProvider
	boshDeployer          boshDeployer
}

func NewDeployBOSHOnAWSForConcourse(infrastructureCreator infrastructureCreator, keyPairSynchronizer keyPairSynchronizer, awsClientProvider awsClientProvider, boshDeployer boshDeployer) DeployBOSHOnAWSForConcourse {
	return DeployBOSHOnAWSForConcourse{
		infrastructureCreator: infrastructureCreator,
		keyPairSynchronizer:   keyPairSynchronizer,
		awsClientProvider:     awsClientProvider,
		boshDeployer:          boshDeployer,
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

	ec2Client, err := d.awsClientProvider.EC2Client(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	if state.KeyPair == nil {
		state.KeyPair = &storage.KeyPair{}
	}

	keyPair, err := d.keyPairSynchronizer.Sync(KeyPair{
		Name:       state.KeyPair.Name,
		PublicKey:  state.KeyPair.PublicKey,
		PrivateKey: state.KeyPair.PrivateKey,
	}, ec2Client)
	if err != nil {
		return state, err
	}

	state.KeyPair.Name = keyPair.Name
	state.KeyPair.PublicKey = keyPair.PublicKey
	state.KeyPair.PrivateKey = keyPair.PrivateKey

	err = d.infrastructureCreator.Create(state.KeyPair.Name, cloudFormationClient)
	if err != nil {
		return state, err
	}

	directorSSLKeyPair := ssl.KeyPair{}
	if state.BOSH != nil {
		directorSSLKeyPair.Certificate = []byte(state.BOSH.DirectorSSLCertificate)
		directorSSLKeyPair.PrivateKey = []byte(state.BOSH.DirectorSSLPrivateKey)
	}

	directorSSLKeyPair, err = d.boshDeployer.Deploy(cloudFormationClient, state.AWS.Region, state.KeyPair.Name, directorSSLKeyPair)
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
