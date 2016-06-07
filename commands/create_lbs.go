package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/multierror"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const CREATE_LBS_COMMAND = "unsupported-create-lbs"

type CreateLBs struct {
	logger                    logger
	certificateManager        certificateManager
	infrastructureManager     infrastructureManager
	boshClientProvider        boshClientProvider
	availabilityZoneRetriever availabilityZoneRetriever
	boshCloudConfigurator     boshCloudConfigurator
	awsCredentialValidator    awsCredentialValidator
	cloudConfigManager        cloudConfigManager
}

type lbConfig struct {
	lbType       string
	certPath     string
	keyPath      string
	chainPath    string
	skipIfExists bool
}

type certificateManager interface {
	Create(certificate, privateKey, chain string) (string, error)
	Describe(certificateName string) (iam.Certificate, error)
	Delete(certificateName string) error
}

type boshClientProvider interface {
	Client(directorAddress, directorUsername, directorPassword string) bosh.Client
}

type boshCloudConfigurator interface {
	Configure(stack cloudformation.Stack, azs []string) bosh.CloudConfigInput
}

func NewCreateLBs(logger logger, awsCredentialValidator awsCredentialValidator, certificateManager certificateManager,
	infrastructureManager infrastructureManager, availabilityZoneRetriever availabilityZoneRetriever,
	boshClientProvider boshClientProvider, boshCloudConfigurator boshCloudConfigurator,
	cloudConfigManager cloudConfigManager) CreateLBs {
	return CreateLBs{
		logger:                    logger,
		certificateManager:        certificateManager,
		infrastructureManager:     infrastructureManager,
		boshClientProvider:        boshClientProvider,
		availabilityZoneRetriever: availabilityZoneRetriever,
		boshCloudConfigurator:     boshCloudConfigurator,
		awsCredentialValidator:    awsCredentialValidator,
		cloudConfigManager:        cloudConfigManager,
	}
}

func (c CreateLBs) Execute(subcommandFlags []string, state storage.State) (storage.State, error) {
	err := c.awsCredentialValidator.Validate()
	if err != nil {
		return state, err
	}

	config, err := c.parseFlags(subcommandFlags)
	if err != nil {
		return state, err
	}

	err = c.validateFlags(config)
	if err != nil {
		return state, err
	}

	if config.skipIfExists && (state.Stack.LBType != "" && state.Stack.LBType != "none") {
		c.logger.Println(fmt.Sprintf("lb type %q exists, skipping...", state.Stack.LBType))
		return state, nil
	}

	boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)

	if err := c.checkFastFails(config.lbType, state.Stack.LBType, state.Stack.Name, boshClient); err != nil {
		return state, err
	}

	certificateName, err := c.certificateManager.Create(config.certPath, config.keyPath, config.chainPath)
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
	lbFlags.String(&config.chainPath, "chain", "")
	lbFlags.Bool(&config.skipIfExists, "skip-if-exists", "", false)

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

	return bblExists(stackName, c.infrastructureManager, boshClient)
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

	cloudConfigInput := c.boshCloudConfigurator.Configure(stack, availabilityZones)

	err = c.cloudConfigManager.Update(cloudConfigInput, boshClient)
	if err != nil {
		return err
	}

	return nil
}

func (c CreateLBs) validateFlags(config lbConfig) error {
	validateErrors := multierror.NewMultiError(CREATE_LBS_COMMAND)
	if config.certPath == "" {
		validateErrors.Add(errors.New("--cert is required"))
	}
	if config.certPath == "" {
		validateErrors.Add(errors.New("--key is required"))
	}

	if validateErrors.Length() > 0 {
		return validateErrors
	}

	return nil
}
