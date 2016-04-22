package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type Destroy struct {
	logger                logger
	stdin                 io.Reader
	boshDeleter           boshDeleter
	clientProvider        clientProvider
	vpcStatusChecker      vpcStatusChecker
	stackManager          stackManager
	stringGenerator       stringGenerator
	infrastructureManager infrastructureManager
	keyPairDeleter        keyPairDeleter
}

type destroyConfig struct {
	NoConfirm bool
}

type keyPairDeleter interface {
	Delete(client ec2.Client, name string) error
}

type boshDeleter interface {
	Delete(boshinit.DeployInput) error
}

type clientProvider interface {
	CloudFormationClient(aws.Config) (cloudformation.Client, error)
	EC2Client(aws.Config) (ec2.Client, error)
}

type vpcStatusChecker interface {
	ValidateSafeToDelete(ec2.Client, string) error
}

type stackManager interface {
	Describe(cloudformation.Client, string) (cloudformation.Stack, error)
}

type stringGenerator interface {
	Generate(prefix string, length int) (string, error)
}

func NewDestroy(logger logger, stdin io.Reader, boshDeleter boshDeleter, clientProvider clientProvider, vpcStatusChecker vpcStatusChecker, stackManager stackManager, stringGenerator stringGenerator, infrastructureManager infrastructureManager, keyPairDeleter keyPairDeleter) Destroy {
	return Destroy{
		logger:                logger,
		stdin:                 stdin,
		boshDeleter:           boshDeleter,
		clientProvider:        clientProvider,
		vpcStatusChecker:      vpcStatusChecker,
		stackManager:          stackManager,
		stringGenerator:       stringGenerator,
		infrastructureManager: infrastructureManager,
		keyPairDeleter:        keyPairDeleter,
	}
}

func (d Destroy) Execute(globalFlags GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	config, err := d.parseFlags(subcommandFlags)
	if err != nil {
		return state, err
	}

	if !config.NoConfirm {
		d.logger.Prompt("Are you sure you want to delete your infrastructure? This operation cannot be undone!")

		var proceed string
		fmt.Fscanln(d.stdin, &proceed)

		proceed = strings.ToLower(proceed)
		if proceed != "yes" && proceed != "y" {
			d.logger.Step("exiting")
			return state, nil
		}
	}

	d.logger.Step("destroying BOSH director and AWS stack")

	cloudFormationClient, ec2Client, err := d.createAWSClients(state, globalFlags)
	if err != nil {
		return state, err
	}

	stack, err := d.stackManager.Describe(cloudFormationClient, state.Stack.Name)
	if err != nil {
		return state, err
	}

	var vpcID = stack.Outputs["VPCID"]
	if err := d.vpcStatusChecker.ValidateSafeToDelete(ec2Client, vpcID); err != nil {
		return state, err
	}

	state, err = d.deleteBOSH(stack, state)
	if err != nil {
		return state, err
	}

	if err := d.infrastructureManager.Delete(cloudFormationClient, state.Stack.Name); err != nil {
		return state, err
	}

	err = d.keyPairDeleter.Delete(ec2Client, state.KeyPair.Name)
	if err != nil {
		return state, err
	}

	return storage.State{}, nil
}

func (d Destroy) parseFlags(subcommandFlags []string) (destroyConfig, error) {
	destroyFlags := flags.New("destroy")

	config := destroyConfig{}
	destroyFlags.Bool(&config.NoConfirm, "n", "no-confirm", false)

	err := destroyFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (d Destroy) createAWSClients(state storage.State, globalFlags GlobalFlags) (cloudformation.Client, ec2.Client, error) {
	awsConfig := aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	}

	cloudFormationClient, err := d.clientProvider.CloudFormationClient(awsConfig)
	if err != nil {
		return nil, nil, err
	}

	ec2Client, err := d.clientProvider.EC2Client(awsConfig)
	if err != nil {
		return nil, nil, err
	}

	return cloudFormationClient, ec2Client, nil
}

func (d Destroy) deleteBOSH(stack cloudformation.Stack, state storage.State) (storage.State, error) {
	infrastructureConfiguration := boshinit.InfrastructureConfiguration{
		AWSRegion:        state.AWS.Region,
		SubnetID:         stack.Outputs["BOSHSubnet"],
		AvailabilityZone: stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        stack.Outputs["BOSHEIP"],
		AccessKeyID:      stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    stack.Outputs["BOSHSecurityGroup"],
	}

	deployInput, err := boshinit.NewDeployInput(state, infrastructureConfiguration, d.stringGenerator)
	if err != nil {
		return state, err
	}

	if err := d.boshDeleter.Delete(deployInput); err != nil {
		return state, err
	}

	return state, nil
}
