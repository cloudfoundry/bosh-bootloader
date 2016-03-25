package unsupported

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type awsClientProvider interface {
	CloudFormationClient(aws.Config) (cloudformation.Client, error)
	EC2Client(aws.Config) (ec2.Client, error)
}

type keyPairSynchronizer interface {
	Sync(keypair ec2.KeyPair, ec2Client ec2.Client) (ec2.KeyPair, error)
}

type infrastructureManager interface {
	Create(keyPairName string, numberOfAZs int, client cloudformation.Client) (cloudformation.Stack, error)
	Exists(client cloudformation.Client) (bool, error)
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
	infrastructureManager     infrastructureManager
	keyPairSynchronizer       keyPairSynchronizer
	awsClientProvider         awsClientProvider
	boshDeployer              boshDeployer
	stringGenerator           stringGenerator
	cloudConfigurator         cloudConfigurator
	availabilityZoneRetriever availabilityZoneRetriever
}

func NewDeployBOSHOnAWSForConcourse(
	infrastructureManager infrastructureManager, keyPairSynchronizer keyPairSynchronizer,
	awsClientProvider awsClientProvider, boshDeployer boshDeployer, stringGenerator stringGenerator,
	cloudConfigurator cloudConfigurator, availabilityZoneRetriever availabilityZoneRetriever) DeployBOSHOnAWSForConcourse {

	return DeployBOSHOnAWSForConcourse{
		infrastructureManager:     infrastructureManager,
		keyPairSynchronizer:       keyPairSynchronizer,
		awsClientProvider:         awsClientProvider,
		boshDeployer:              boshDeployer,
		stringGenerator:           stringGenerator,
		cloudConfigurator:         cloudConfigurator,
		availabilityZoneRetriever: availabilityZoneRetriever,
	}
}

func (d DeployBOSHOnAWSForConcourse) Execute(globalFlags commands.GlobalFlags, state storage.State) (storage.State, error) {
	cloudformationClient, err := d.awsClientProvider.CloudFormationClient(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	stackExists, err := d.infrastructureManager.Exists(cloudformationClient)
	if err != nil {
		return state, err
	}

	if state.BOSH != nil && !stackExists {
		return state, fmt.Errorf(
			"Found BOSH data in state directory, but Cloud Formation stack %q cannot be found "+
				"for region %q and given AWS credentials. bbl cannot safely proceed. Open an issue on GitHub at "+
				"https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you need assistance.",
			cloudformation.STACK_NAME, state.AWS.Region)
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

	stack, err := d.infrastructureManager.Create(state.KeyPair.Name, len(availabilityZones), cloudformationClient)
	if err != nil {
		return state, err
	}

	boshDeployInput, err := boshinit.NewBOSHDeployInput(state, stack, d.stringGenerator)
	if err != nil {
		return state, err
	}

	boshInitOutput, err := d.boshDeployer.Deploy(boshDeployInput)
	if err != nil {
		return state, err
	}

	if state.BOSH == nil {
		state.BOSH = &storage.BOSH{
			DirectorUsername:       boshDeployInput.DirectorUsername,
			DirectorPassword:       boshDeployInput.DirectorPassword,
			DirectorSSLCertificate: string(boshInitOutput.DirectorSSLKeyPair.Certificate),
			DirectorSSLPrivateKey:  string(boshInitOutput.DirectorSSLKeyPair.PrivateKey),
			Credentials:            boshInitOutput.Credentials,
			State:                  boshInitOutput.BOSHInitState,
		}
	}

	err = d.cloudConfigurator.Configure(stack, availabilityZones)
	if err != nil {
		return state, err
	}

	return state, nil
}
