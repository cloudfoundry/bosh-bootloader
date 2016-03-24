package unsupported

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const USERNAME_PREFIX = "user-"
const USERNAME_LENGTH = 7
const PASSWORD_PREFIX = "p-"
const PASSWORD_LENGTH = 15

type awsClientProvider interface {
	CloudFormationClient(aws.Config) (cloudformation.Client, error)
	EC2Client(aws.Config) (ec2.Client, error)
}

type keyPairSynchronizer interface {
	Sync(keypair ec2.KeyPair, ec2Client ec2.Client) (ec2.KeyPair, error)
}

type infrastructureCreator interface {
	Create(keyPairName string, numberOfAZs int, client cloudformation.Client) (cloudformation.Stack, error)
}

type boshDeployer interface {
	Deploy(boshinit.BOSHDeployInput) (boshinit.BOSHDeployOutput, error)
}

type stringGenerator interface {
	Generate(string, int) (string, error)
}

type cloudConfigurator interface {
	Configure(stack cloudformation.Stack, azs []string) error
}

type availabilityZoneRetriever interface {
	Retrieve(region string, client ec2.Client) ([]string, error)
}

type logger interface {
	Step(string)
	Println(string)
}

type DeployBOSHOnAWSForConcourse struct {
	infrastructureCreator     infrastructureCreator
	keyPairSynchronizer       keyPairSynchronizer
	awsClientProvider         awsClientProvider
	boshDeployer              boshDeployer
	stringGenerator           stringGenerator
	cloudConfigurator         cloudConfigurator
	availabilityZoneRetriever availabilityZoneRetriever
}

func NewDeployBOSHOnAWSForConcourse(
	infrastructureCreator infrastructureCreator, keyPairSynchronizer keyPairSynchronizer,
	awsClientProvider awsClientProvider, boshDeployer boshDeployer, stringGenerator stringGenerator,
	cloudConfigurator cloudConfigurator, availabilityZoneRetriever availabilityZoneRetriever) DeployBOSHOnAWSForConcourse {

	return DeployBOSHOnAWSForConcourse{
		infrastructureCreator:     infrastructureCreator,
		keyPairSynchronizer:       keyPairSynchronizer,
		awsClientProvider:         awsClientProvider,
		boshDeployer:              boshDeployer,
		stringGenerator:           stringGenerator,
		cloudConfigurator:         cloudConfigurator,
		availabilityZoneRetriever: availabilityZoneRetriever,
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

	keyPair, err := d.keyPairSynchronizer.Sync(ec2.KeyPair{
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

	availabilityZones, err := d.availabilityZoneRetriever.Retrieve(state.AWS.Region, ec2Client)
	if err != nil {
		return state, err
	}

	stack, err := d.infrastructureCreator.Create(state.KeyPair.Name, len(availabilityZones), cloudFormationClient)
	if err != nil {
		return state, err
	}

	boshOutput := boshinit.BOSHDeployOutput{
		DirectorSSLKeyPair: ssl.KeyPair{},
		BOSHInitState:      boshinit.State{},
		Credentials:        manifests.InternalCredentials{},
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

	boshOutput, err = d.boshDeployer.Deploy(boshinit.BOSHDeployInput{
		DirectorUsername: directorUsername,
		DirectorPassword: directorPassword,
		State:            boshOutput.BOSHInitState,
		Stack:            stack,
		AWSRegion:        state.AWS.Region,
		SSLKeyPair:       boshOutput.DirectorSSLKeyPair,
		EC2KeyPair: ec2.KeyPair{
			Name:       state.KeyPair.Name,
			PrivateKey: state.KeyPair.PrivateKey,
			PublicKey:  state.KeyPair.PublicKey,
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

	err = d.cloudConfigurator.Configure(stack, availabilityZones)
	if err != nil {
		return state, err
	}

	return state, nil
}
