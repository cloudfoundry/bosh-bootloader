package commands

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type CreateLBs struct {
	clientProvider            awsClientProvider
	certificateManager        certificateManager
	infrastructureManager     infrastructureManager
	boshClientProvider        boshClientProvider
	availabilityZoneRetriever availabilityZoneRetriever
	boshCloudConfigurator     boshCloudConfigurator
}

type lbConfig struct {
	lbType   string
	certPath string
	keyPath  string
}

type certificateManager interface {
	Create(certificate, privateKey string, client iam.Client) (string, error)
	Describe(certificateName string, client iam.Client) (iam.Certificate, error)
	Delete(certificateName string, client iam.Client) error
}

type boshClientProvider interface {
	Client(directorAddress, directorUsername, directorPassword string) bosh.Client
}

type boshCloudConfigurator interface {
	Configure(stack cloudformation.Stack, azs []string, client bosh.Client) error
}

func NewCreateLBs(clientProvider awsClientProvider, certificateManager certificateManager,
	infrastructureManager infrastructureManager, availabilityZoneRetriever availabilityZoneRetriever,
	boshClientProvider boshClientProvider, boshCloudConfigurator boshCloudConfigurator) CreateLBs {
	return CreateLBs{
		clientProvider:            clientProvider,
		certificateManager:        certificateManager,
		infrastructureManager:     infrastructureManager,
		boshClientProvider:        boshClientProvider,
		availabilityZoneRetriever: availabilityZoneRetriever,
		boshCloudConfigurator:     boshCloudConfigurator,
	}
}

func (c CreateLBs) Execute(globalFlags GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	config, err := c.parseFlags(subcommandFlags)
	if err != nil {
		return state, err
	}

	awsConfig := aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	}

	cloudFormationClient, err := c.clientProvider.CloudFormationClient(awsConfig)
	if err != nil {
		return state, err
	}

	iamClient, err := c.clientProvider.IAMClient(awsConfig)
	if err != nil {
		return state, err
	}

	ec2Client, err := c.clientProvider.EC2Client(awsConfig)
	if err != nil {
		return state, err
	}

	boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)

	if err := c.checkFastFails(config.lbType, state.Stack.LBType, state.Stack.Name, boshClient, cloudFormationClient); err != nil {
		return state, err
	}

	certificateName, err := c.certificateManager.Create(config.certPath, config.keyPath, iamClient)
	if err != nil {
		return state, err
	}

	state.CertificateName = certificateName
	state.Stack.LBType = config.lbType

	if err := c.updateStackAndBOSH(awsConfig.Region, certificateName, state.KeyPair.Name, state.Stack.Name, config.lbType, ec2Client, iamClient, cloudFormationClient, boshClient); err != nil {
		return state, err
	}

	return state, nil
}

func (CreateLBs) parseFlags(subcommandFlags []string) (lbConfig, error) {
	lbFlags := flags.New("create-lbs")

	config := lbConfig{}
	lbFlags.String(&config.lbType, "type", "")
	lbFlags.String(&config.certPath, "cert", "")
	lbFlags.String(&config.keyPath, "key", "")

	err := lbFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (CreateLBs) isValidLBType(lbType string) bool {
	for _, v := range []string{"concourse", "cf"} {
		if lbType == v {
			return true
		}
	}

	return false
}

func (c CreateLBs) checkFastFails(newLBType string, currentLBType string, stackName string, boshClient bosh.Client, cloudFormationClient cloudformation.Client) error {
	if !c.isValidLBType(newLBType) {
		return fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse and cf", newLBType)
	}

	if currentLBType == "concourse" || currentLBType == "cf" {
		return fmt.Errorf("bbl already has a %s load balancer attached, please remove the previous load balancer before attaching a new one", currentLBType)
	}

	_, err := c.infrastructureManager.Describe(cloudFormationClient, stackName)
	if err != nil {
		return err
	}

	_, err = boshClient.Info()
	if err != nil {
		return fmt.Errorf("bosh director cannot be reached: %s", err.Error())
	}
	return nil
}

func (c CreateLBs) updateStackAndBOSH(
	awsRegion string, certificateName string, keyPairName string, stackName string, lbType string,
	ec2Client ec2.Client, iamClient iam.Client, cloudFormationClient cloudformation.Client, boshClient bosh.Client,
) error {

	availabilityZones, err := c.availabilityZoneRetriever.Retrieve(awsRegion, ec2Client)
	if err != nil {
		return err
	}

	certificate, err := c.certificateManager.Describe(certificateName, iamClient)

	stack, err := c.infrastructureManager.Update(keyPairName, len(availabilityZones), stackName, lbType, certificate.ARN, cloudFormationClient)
	if err != nil {
		return err
	}

	err = c.boshCloudConfigurator.Configure(stack, availabilityZones, boshClient)
	if err != nil {
		return err
	}

	return nil
}
