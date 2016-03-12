package unsupported

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const STACKNAME = "concourse"

type awsClientProvider interface {
	CloudFormationClient(aws.Config) (cloudformation.Client, error)
	EC2Client(aws.Config) (ec2.Client, error)
}

type keyPairSynchronizer interface {
	Sync(keypair KeyPair, ec2Client ec2.Client) (KeyPair, error)
}

type infrastructureCreator interface {
	Create(keyPairName string, client cloudformation.Client) (cloudformation.Stack, error)
}

type boshDeployer interface {
	Deploy(BOSHDeployInput) (boshinit.State, ssl.KeyPair, error)
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

	stack, err := d.infrastructureCreator.Create(state.KeyPair.Name, cloudFormationClient)
	if err != nil {
		return state, err
	}

	directorSSLKeyPair := ssl.KeyPair{}
	boshInitState := boshinit.State{}
	if state.BOSH != nil {
		directorSSLKeyPair.Certificate = []byte(state.BOSH.DirectorSSLCertificate)
		directorSSLKeyPair.PrivateKey = []byte(state.BOSH.DirectorSSLPrivateKey)
		if state.BOSH.State != nil {
			boshInitState = state.BOSH.State
		}
	}

	boshInitState, directorSSLKeyPair, err = d.boshDeployer.Deploy(BOSHDeployInput{
		State:      boshInitState,
		Stack:      stack,
		AWSRegion:  state.AWS.Region,
		SSLKeyPair: directorSSLKeyPair,
		EC2KeyPair: ec2.KeyPair{
			Name:       state.KeyPair.Name,
			PrivateKey: []byte(state.KeyPair.PrivateKey),
			PublicKey:  []byte(state.KeyPair.PublicKey),
		},
	})
	if err != nil {
		return state, err
	}

	if state.BOSH == nil {
		state.BOSH = &storage.BOSH{
			DirectorSSLCertificate: string(directorSSLKeyPair.Certificate),
			DirectorSSLPrivateKey:  string(directorSSLKeyPair.PrivateKey),
			State: boshInitState,
		}
	}

	return state, nil
}
