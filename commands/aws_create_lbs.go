package commands

import (
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSCreateLBs struct {
	logger                    logger
	certificateManager        certificateManager
	infrastructureManager     infrastructureManager
	availabilityZoneRetriever availabilityZoneRetriever
	credentialValidator       credentialValidator
	cloudConfigManager        cloudConfigManager
	certificateValidator      certificateValidator
	guidGenerator             guidGenerator
	stateStore                stateStore
	stateValidator            stateValidator
	terraformManager          terraformManager
	environmentValidator      environmentValidator
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

type certificateValidator interface {
	Validate(command, certPath, keyPath, chainPath string) error
}

type environmentValidator interface {
	Validate(state storage.State) error
}

type guidGenerator interface {
	Generate() (string, error)
}

func NewAWSCreateLBs(logger logger, credentialValidator credentialValidator, certificateManager certificateManager,
	infrastructureManager infrastructureManager, availabilityZoneRetriever availabilityZoneRetriever,
	cloudConfigManager cloudConfigManager, certificateValidator certificateValidator,
	guidGenerator guidGenerator, stateStore stateStore, terraformManager terraformManager, environmentValidator environmentValidator) AWSCreateLBs {
	return AWSCreateLBs{
		logger:                    logger,
		certificateManager:        certificateManager,
		infrastructureManager:     infrastructureManager,
		availabilityZoneRetriever: availabilityZoneRetriever,
		credentialValidator:       credentialValidator,
		cloudConfigManager:        cloudConfigManager,
		certificateValidator:      certificateValidator,
		guidGenerator:             guidGenerator,
		stateStore:                stateStore,
		terraformManager:          terraformManager,
		environmentValidator:      environmentValidator,
	}
}

func (c AWSCreateLBs) Execute(config AWSCreateLBsConfig, state storage.State) error {
	err := c.credentialValidator.Validate()
	if err != nil {
		return err
	}

	if config.SkipIfExists && lbExists(state.Stack.LBType) {
		c.logger.Println(fmt.Sprintf("lb type %q exists, skipping...", state.Stack.LBType))
		return nil
	}

	if err := c.checkFastFails(config.LBType, state.Stack.LBType); err != nil {
		return err
	}

	if err := c.environmentValidator.Validate(state); err != nil {
		return err
	}

	if state.TFState != "" {
		if config.LBType == "cf" {
			certContents, err := ioutil.ReadFile(config.CertPath)
			if err != nil {
				return err
			}
			keyContents, err := ioutil.ReadFile(config.KeyPath)
			if err != nil {
				return err
			}
			state.LB.Cert = string(certContents)
			state.LB.Key = string(keyContents)
		}

		state.LB.Type = config.LBType

		state, err = c.terraformManager.Apply(state)
		if err != nil {
			return err
		}
	} else {
		err = c.certificateValidator.Validate(CreateLBsCommand, config.CertPath, config.KeyPath, config.ChainPath)
		if err != nil {
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

		if err := c.updateStack(state.AWS.Region, certificateName, state.KeyPair.Name, state.Stack.Name, state.Stack.BOSHAZ, config.LBType, state.EnvID); err != nil {
			return err
		}
	}

	err = c.stateStore.Set(state)
	if err != nil {
		return err
	}

	if !state.NoDirector {
		err = c.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	return nil
}

func (AWSCreateLBs) isValidLBType(lbType string) bool {
	return lbType == "concourse" || lbType == "cf"
}

func (c AWSCreateLBs) checkFastFails(newLBType string, currentLBType string) error {
	if newLBType == "" {
		return fmt.Errorf("--type is a required flag")
	}

	if !c.isValidLBType(newLBType) {
		return fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse and cf", newLBType)
	}

	if lbExists(currentLBType) {
		return fmt.Errorf("bbl already has a %s load balancer attached, please remove the previous load balancer before attaching a new one", currentLBType)
	}

	return nil
}

func (c AWSCreateLBs) updateStack(awsRegion string, certificateName string, keyPairName string, stackName string, boshAZ, lbType string,
	envID string) error {
	availabilityZones, err := c.availabilityZoneRetriever.Retrieve(awsRegion)
	if err != nil {
		return err
	}

	certificate, err := c.certificateManager.Describe(certificateName)

	_, err = c.infrastructureManager.Update(keyPairName, availabilityZones, stackName, boshAZ, lbType, certificate.ARN, envID)
	if err != nil {
		return err
	}

	return nil
}
