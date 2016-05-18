package commands

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type CreateLBs struct {
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
	Create(certificate, privateKey string) (string, error)
	Describe(certificateName string) (iam.Certificate, error)
	Delete(certificateName string) error
}

type boshClientProvider interface {
	Client(directorAddress, directorUsername, directorPassword string) bosh.Client
}

type boshCloudConfigurator interface {
	Configure(stack cloudformation.Stack, azs []string, client bosh.Client) error
}

func NewCreateLBs(certificateManager certificateManager, infrastructureManager infrastructureManager,
	availabilityZoneRetriever availabilityZoneRetriever, boshClientProvider boshClientProvider,
	boshCloudConfigurator boshCloudConfigurator) CreateLBs {
	return CreateLBs{
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

	boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)

	if err := c.checkFastFails(config.lbType, state.Stack.LBType, state.Stack.Name, boshClient); err != nil {
		return state, err
	}

	certificateName, err := c.certificateManager.Create(config.certPath, config.keyPath)
	if err != nil {
		return state, err
	}

	state.Stack.CertificateName = certificateName
	state.Stack.LBType = config.lbType

	if err := c.updateStackAndBOSH(state.AWS.Region, certificateName, state.KeyPair.Name, state.Stack.Name, config.lbType, boshClient); err != nil {
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

func (c CreateLBs) checkFastFails(newLBType string, currentLBType string, stackName string, boshClient bosh.Client) error {
	if !c.isValidLBType(newLBType) {
		return fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse and cf", newLBType)
	}

	if currentLBType == "concourse" || currentLBType == "cf" {
		return fmt.Errorf("bbl already has a %s load balancer attached, please remove the previous load balancer before attaching a new one", currentLBType)
	}

	if stackExists, err := c.infrastructureManager.Exists(stackName); err != nil {
		return err
	} else if !stackExists {
		return BBLNotFound
	}

	if _, err := boshClient.Info(); err != nil {
		return BBLNotFound
	}

	return nil
}

func (c CreateLBs) updateStackAndBOSH(
	awsRegion string, certificateName string, keyPairName string, stackName string,
	lbType string, boshClient bosh.Client,
) error {

	availabilityZones, err := c.availabilityZoneRetriever.Retrieve(awsRegion)
	if err != nil {
		return err
	}

	certificate, err := c.certificateManager.Describe(certificateName)

	stack, err := c.infrastructureManager.Update(keyPairName, len(availabilityZones), stackName, lbType, certificate.ARN)
	if err != nil {
		return err
	}

	err = c.boshCloudConfigurator.Configure(stack, availabilityZones, boshClient)
	if err != nil {
		return err
	}

	return nil
}
