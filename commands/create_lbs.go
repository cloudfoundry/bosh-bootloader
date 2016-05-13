package commands

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
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

	if !c.isValidLBType(config.lbType) {
		return state, fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse and cf", config.lbType)
	}

	if state.Stack.LBType == "concourse" || state.Stack.LBType == "cf" {
		return state, fmt.Errorf("bbl already has a %s load balancer attached, please remove the previous load balancer before attaching a new one", state.Stack.LBType)
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

	_, err = c.infrastructureManager.Describe(cloudFormationClient, state.Stack.Name)
	if err != nil {
		return state, err
	}

	boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)

	_, err = boshClient.Info()
	if err != nil {
		return state, fmt.Errorf("bosh director cannot be reached: %s", err.Error())
	}

	iamClient, err := c.clientProvider.IAMClient(awsConfig)
	if err != nil {
		return state, err
	}

	certificateName, err := c.certificateManager.Create(config.certPath, config.keyPath, iamClient)
	if err != nil {
		return state, err
	}
	state.CertificateName = certificateName
	state.Stack.LBType = config.lbType

	ec2Client, err := c.clientProvider.EC2Client(awsConfig)
	if err != nil {
		return state, err
	}

	availabilityZones, err := c.availabilityZoneRetriever.Retrieve(awsConfig.Region, ec2Client)
	if err != nil {
		return state, err
	}

	certificate, err := c.certificateManager.Describe(certificateName, iamClient)

	stack, err := c.infrastructureManager.Update(state.KeyPair.Name, len(availabilityZones), state.Stack.Name, config.lbType, certificate.ARN, cloudFormationClient)
	if err != nil {
		return state, err
	}

	err = c.boshCloudConfigurator.Configure(stack, availabilityZones, boshClient)
	if err != nil {
		return state, err
	}

	return state, nil
}

func (c CreateLBs) parseFlags(subcommandFlags []string) (lbConfig, error) {
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

func (c CreateLBs) isValidLBType(lbType string) bool {
	for _, v := range []string{"concourse", "cf"} {
		if lbType == v {
			return true
		}
	}

	return false
}
