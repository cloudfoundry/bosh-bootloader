package commands

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
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
	Create(keyPairName string, numberOfAZs int, stackName string, lbType string, client cloudformation.Client) (cloudformation.Stack, error)
	Exists(stackName string, client cloudformation.Client) (bool, error)
	Delete(client cloudformation.Client, stackName string) error
}

type boshDeployer interface {
	Deploy(boshinit.DeployInput) (boshinit.DeployOutput, error)
}

type cloudConfigurator interface {
	Configure(stack cloudformation.Stack, azs []string, boshClient bosh.Client) error
}

type availabilityZoneRetriever interface {
	Retrieve(region string, client ec2.Client) ([]string, error)
}

type logger interface {
	Step(string)
	Println(string)
	Prompt(string)
}

type upConfig struct {
	lbType string
}

type Up struct {
	infrastructureManager     infrastructureManager
	keyPairSynchronizer       keyPairSynchronizer
	awsClientProvider         awsClientProvider
	boshDeployer              boshDeployer
	stringGenerator           stringGenerator
	cloudConfigurator         cloudConfigurator
	availabilityZoneRetriever availabilityZoneRetriever
}

func NewUp(
	infrastructureManager infrastructureManager, keyPairSynchronizer keyPairSynchronizer,
	awsClientProvider awsClientProvider, boshDeployer boshDeployer, stringGenerator stringGenerator,
	cloudConfigurator cloudConfigurator, availabilityZoneRetriever availabilityZoneRetriever) Up {

	return Up{
		infrastructureManager:     infrastructureManager,
		keyPairSynchronizer:       keyPairSynchronizer,
		awsClientProvider:         awsClientProvider,
		boshDeployer:              boshDeployer,
		stringGenerator:           stringGenerator,
		cloudConfigurator:         cloudConfigurator,
		availabilityZoneRetriever: availabilityZoneRetriever,
	}
}

func (Up) parseFlags(subcommandFlags []string) (upConfig, error) {
	upFlags := flags.New("unsupported-deploy-bosh-on-aws-for-concourse")

	config := upConfig{}
	upFlags.String(&config.lbType, "lb-type", "")

	err := upFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (u Up) Execute(globalFlags GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	config, err := u.parseFlags(subcommandFlags)
	if err != nil {
		return state, err
	}

	cloudFormationClient, err := u.awsClientProvider.CloudFormationClient(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	stackExists, err := u.infrastructureManager.Exists(state.Stack.Name, cloudFormationClient)
	if err != nil {
		return state, err
	}

	if !state.BOSH.IsEmpty() && !stackExists {
		return state, fmt.Errorf(
			"Found BOSH data in state directory, but Cloud Formation stack %q cannot be found "+
				"for region %q and given AWS credentials. bbl cannot safely proceed. Open an issue on GitHub at "+
				"https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you need assistance.",
			state.Stack.Name, state.AWS.Region)
	}

	ec2Client, err := u.awsClientProvider.EC2Client(aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	})
	if err != nil {
		return state, err
	}

	keyPair, err := u.keyPairSynchronizer.Sync(ec2.KeyPair{
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

	availabilityZones, err := u.availabilityZoneRetriever.Retrieve(state.AWS.Region, ec2Client)
	if err != nil {
		return state, err
	}

	if state.Stack.Name == "" {
		state.Stack.Name, err = u.stringGenerator.Generate("bbl-aws-", 5)
		if err != nil {
			return state, err
		}
	}

	stack, err := u.infrastructureManager.Create(state.KeyPair.Name, len(availabilityZones), state.Stack.Name, config.lbType, cloudFormationClient)
	if err != nil {
		return state, err
	}

	infrastructureConfiguration := boshinit.InfrastructureConfiguration{
		AWSRegion:        state.AWS.Region,
		SubnetID:         stack.Outputs["BOSHSubnet"],
		AvailabilityZone: stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        stack.Outputs["BOSHEIP"],
		AccessKeyID:      stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    stack.Outputs["BOSHSecurityGroup"],
	}

	deployInput, err := boshinit.NewDeployInput(state, infrastructureConfiguration, u.stringGenerator)
	if err != nil {
		return state, err
	}

	deployOutput, err := u.boshDeployer.Deploy(deployInput)
	if err != nil {
		return state, err
	}

	if state.BOSH.IsEmpty() {
		state.BOSH = storage.BOSH{
			DirectorAddress:        stack.Outputs["BOSHURL"],
			DirectorUsername:       deployInput.DirectorUsername,
			DirectorPassword:       deployInput.DirectorPassword,
			DirectorSSLCertificate: string(deployOutput.DirectorSSLKeyPair.Certificate),
			DirectorSSLPrivateKey:  string(deployOutput.DirectorSSLKeyPair.PrivateKey),
			Credentials:            deployOutput.Credentials,
			State:                  deployOutput.BOSHInitState,
			Manifest:               deployOutput.BOSHInitManifest,
		}
	}

	err = u.cloudConfigurator.Configure(stack, availabilityZones, bosh.NewClient(stack.Outputs["BOSHURL"], deployInput.DirectorUsername, deployInput.DirectorPassword))
	if err != nil {
		return state, err
	}

	return state, nil
}
