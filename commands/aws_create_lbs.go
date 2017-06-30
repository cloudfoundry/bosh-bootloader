package commands

import (
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSCreateLBs struct {
	logger               logger
	credentialValidator  credentialValidator
	cloudConfigManager   cloudConfigManager
	certificateValidator certificateValidator
	stateStore           stateStore
	stateValidator       stateValidator
	terraformManager     terraformApplier
	environmentValidator environmentValidator
}

type AWSCreateLBsConfig struct {
	LBType       string
	CertPath     string
	KeyPath      string
	ChainPath    string
	Domain       string
	SkipIfExists bool
}

type certificateValidator interface {
	Validate(command, certPath, keyPath, chainPath string) error
}

type environmentValidator interface {
	Validate(state storage.State) error
}

func NewAWSCreateLBs(logger logger, credentialValidator credentialValidator,
	cloudConfigManager cloudConfigManager, certificateValidator certificateValidator,
	stateStore stateStore, terraformManager terraformApplier, environmentValidator environmentValidator) AWSCreateLBs {
	return AWSCreateLBs{
		logger:               logger,
		credentialValidator:  credentialValidator,
		cloudConfigManager:   cloudConfigManager,
		certificateValidator: certificateValidator,
		stateStore:           stateStore,
		terraformManager:     terraformManager,
		environmentValidator: environmentValidator,
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

	err = c.certificateValidator.Validate(CreateLBsCommand, config.CertPath, config.KeyPath, config.ChainPath)
	if err != nil {
		return err
	}

	if config.LBType == "cf" || config.LBType == "concourse" {
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

		if config.ChainPath != "" {
			chainContents, err := ioutil.ReadFile(config.ChainPath)
			if err != nil {
				return err
			}

			state.LB.Chain = string(chainContents)
		}
	}

	if config.Domain != "" {
		state.LB.Domain = config.Domain
	}

	state.LB.Type = config.LBType

	err = c.stateStore.Set(state)
	if err != nil {
		return err
	}

	state, err = c.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, c.stateStore)
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
