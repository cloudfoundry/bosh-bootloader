package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSCreateLBs struct {
	logger                    logger
	certificateManager        certificateManager
	infrastructureManager     infrastructureManager
	boshClientProvider        boshClientProvider
	availabilityZoneRetriever availabilityZoneRetriever
	boshCloudConfigurator     boshCloudConfigurator
	credentialValidator       credentialValidator
	cloudConfigManager        cloudConfigManager
	certificateValidator      certificateValidator
	guidGenerator             guidGenerator
	stateStore                stateStore
	stateValidator            stateValidator
}

type AWSCreateLBsConfig struct {
	LBType       string
	CertPath     string
	KeyPath      string
	ChainPath    string
	SkipIfExists bool
}

type certificateManager interface {
	Create(certificate, privateKey, chain, certificateName string) error
	Describe(certificateName string) (iam.Certificate, error)
	Delete(certificateName string) error
}

type boshClientProvider interface {
	Client(directorAddress, directorUsername, directorPassword string) bosh.Client
}

type boshCloudConfigurator interface {
	Configure(stack cloudformation.Stack, azs []string) bosh.CloudConfigInput
}

type certificateValidator interface {
	Validate(command, certPath, keyPath, chainPath string) error
}

type guidGenerator interface {
	Generate() (string, error)
}

func NewAWSCreateLBs(logger logger, credentialValidator credentialValidator, certificateManager certificateManager,
	infrastructureManager infrastructureManager, availabilityZoneRetriever availabilityZoneRetriever, boshClientProvider boshClientProvider,
	boshCloudConfigurator boshCloudConfigurator, cloudConfigManager cloudConfigManager, certificateValidator certificateValidator,
	guidGenerator guidGenerator, stateStore stateStore) AWSCreateLBs {
	return AWSCreateLBs{
		logger:                    logger,
		certificateManager:        certificateManager,
		infrastructureManager:     infrastructureManager,
		boshClientProvider:        boshClientProvider,
		availabilityZoneRetriever: availabilityZoneRetriever,
		boshCloudConfigurator:     boshCloudConfigurator,
		credentialValidator:       credentialValidator,
		cloudConfigManager:        cloudConfigManager,
		certificateValidator:      certificateValidator,
		guidGenerator:             guidGenerator,
		stateStore:                stateStore,
	}
}

func (c AWSCreateLBs) Execute(config AWSCreateLBsConfig, state storage.State) error {
	err := c.credentialValidator.ValidateAWS()
	if err != nil {
		return err
	}

	err = c.certificateValidator.Validate(CreateLBsCommand, config.CertPath, config.KeyPath, config.ChainPath)
	if err != nil {
		return err
	}

	if config.SkipIfExists && lbExists(state.Stack.LBType) {
		c.logger.Println(fmt.Sprintf("lb type %q exists, skipping...", state.Stack.LBType))
		return nil
	}

	boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)

	if err := c.checkFastFails(config.LBType, state.Stack.LBType, state.Stack.Name, boshClient); err != nil {
		return err
	}

	c.logger.Step("uploading certificate")

	certificateName, err := certificateNameFor(config.LBType, c.guidGenerator, state.EnvID)
	if err != nil {
		return err
	}

	err = c.certificateManager.Create(config.CertPath, config.KeyPath, config.ChainPath, certificateName)
	if err != nil {
		return err
	}

	state.Stack.CertificateName = certificateName
	state.Stack.LBType = config.LBType

	if err := c.updateStackAndBOSH(state.AWS.Region, certificateName, state.KeyPair.Name, state.Stack.Name, config.LBType, boshClient, state.EnvID); err != nil {
		return err
	}

	err = c.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}

func (AWSCreateLBs) isValidLBType(lbType string) bool {
	return lbType == "concourse" || lbType == "cf"
}

func (c AWSCreateLBs) checkFastFails(newLBType string, currentLBType string, stackName string, boshClient bosh.Client) error {
	if newLBType == "" {
		return fmt.Errorf("--type is a required flag")
	}

	if !c.isValidLBType(newLBType) {
		return fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse and cf", newLBType)
	}

	if lbExists(currentLBType) {
		return fmt.Errorf("bbl already has a %s load balancer attached, please remove the previous load balancer before attaching a new one", currentLBType)
	}

	return bblExists(stackName, c.infrastructureManager, boshClient)
}

func (c AWSCreateLBs) updateStackAndBOSH(
	awsRegion string, certificateName string, keyPairName string, stackName string,
	lbType string, boshClient bosh.Client, envID string,
) error {

	availabilityZones, err := c.availabilityZoneRetriever.Retrieve(awsRegion)
	if err != nil {
		return err
	}

	certificate, err := c.certificateManager.Describe(certificateName)

	stack, err := c.infrastructureManager.Update(keyPairName, availabilityZones, stackName, lbType, certificate.ARN, envID)
	if err != nil {
		return err
	}

	cloudConfigInput := c.boshCloudConfigurator.Configure(stack, availabilityZones)

	err = c.cloudConfigManager.Update(cloudConfigInput, boshClient)
	if err != nil {
		return err
	}

	return nil
}
