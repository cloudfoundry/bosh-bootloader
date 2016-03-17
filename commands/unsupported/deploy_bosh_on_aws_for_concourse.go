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
const USERNAME_PREFIX = "user-"
const USERNAME_LENGTH = 7
const PASSWORD_PREFIX = "p-"
const PASSWORD_LENGTH = 15

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
	Deploy(BOSHDeployInput) (BOSHDeployOutput, error)
}

type stringGenerator interface {
	Generate(string, int) (string, error)
}

type DeployBOSHOnAWSForConcourse struct {
	infrastructureCreator infrastructureCreator
	keyPairSynchronizer   keyPairSynchronizer
	awsClientProvider     awsClientProvider
	boshDeployer          boshDeployer
	stringGenerator       stringGenerator
}

func NewDeployBOSHOnAWSForConcourse(infrastructureCreator infrastructureCreator, keyPairSynchronizer keyPairSynchronizer, awsClientProvider awsClientProvider, boshDeployer boshDeployer, stringGenerator stringGenerator) DeployBOSHOnAWSForConcourse {
	return DeployBOSHOnAWSForConcourse{
		infrastructureCreator: infrastructureCreator,
		keyPairSynchronizer:   keyPairSynchronizer,
		awsClientProvider:     awsClientProvider,
		boshDeployer:          boshDeployer,
		stringGenerator:       stringGenerator,
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

	boshOutput := BOSHDeployOutput{
		DirectorSSLKeyPair: ssl.KeyPair{},
		BOSHInitState:      boshinit.State{},
		Credentials:        boshinit.InternalCredentials{},
	}
	var directorUsername string
	var directorPassword string
	if state.BOSH != nil {
		boshOutput.DirectorSSLKeyPair.Certificate = []byte(state.BOSH.DirectorSSLCertificate)
		boshOutput.DirectorSSLKeyPair.PrivateKey = []byte(state.BOSH.DirectorSSLPrivateKey)
		boshOutput.Credentials = state.BOSH.Credentials
		if state.BOSH.State != nil {
			boshOutput.BOSHInitState = state.BOSH.State
		}
		directorUsername = state.BOSH.DirectorUsername
		directorPassword = state.BOSH.DirectorPassword
	}

	if directorUsername == "" {
		directorUsername, err = d.stringGenerator.Generate(USERNAME_PREFIX, USERNAME_LENGTH)
		if err != nil {
			return state, err
		}
	}

	if directorPassword == "" {
		directorPassword, err = d.stringGenerator.Generate(PASSWORD_PREFIX, PASSWORD_LENGTH)
		if err != nil {
			return state, err
		}
	}

	boshOutput, err = d.boshDeployer.Deploy(BOSHDeployInput{
		DirectorUsername: directorUsername,
		DirectorPassword: directorPassword,
		State:            boshOutput.BOSHInitState,
		Stack:            stack,
		AWSRegion:        state.AWS.Region,
		SSLKeyPair:       boshOutput.DirectorSSLKeyPair,
		EC2KeyPair: ec2.KeyPair{
			Name:       state.KeyPair.Name,
			PrivateKey: []byte(state.KeyPair.PrivateKey),
			PublicKey:  []byte(state.KeyPair.PublicKey),
		},
		Credentials: boshOutput.Credentials,
	})
	if err != nil {
		return state, err
	}

	if state.BOSH == nil {
		state.BOSH = &storage.BOSH{
			DirectorUsername:       directorUsername,
			DirectorPassword:       directorPassword,
			DirectorSSLCertificate: string(boshOutput.DirectorSSLKeyPair.Certificate),
			DirectorSSLPrivateKey:  string(boshOutput.DirectorSSLKeyPair.PrivateKey),
			Credentials:            boshOutput.Credentials,
			State:                  boshOutput.BOSHInitState,
		}
	}

	return state, nil
}
