package commands

import (
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSCreateLBs struct {
	cloudConfigManager   cloudConfigManager
	stateStore           stateStore
	stateValidator       stateValidator
	terraformManager     terraformApplier
	environmentValidator EnvironmentValidator
}

type AWSCreateLBsConfig struct {
	LBType    string
	CertPath  string
	KeyPath   string
	ChainPath string
	Domain    string
}

type EnvironmentValidator interface {
	Validate(state storage.State) error
}

func NewAWSCreateLBs(cloudConfigManager cloudConfigManager, stateStore stateStore,
	terraformManager terraformApplier, environmentValidator EnvironmentValidator) AWSCreateLBs {
	return AWSCreateLBs{
		cloudConfigManager:   cloudConfigManager,
		stateStore:           stateStore,
		terraformManager:     terraformManager,
		environmentValidator: environmentValidator,
	}
}

func (c AWSCreateLBs) Execute(config CreateLBsConfig, state storage.State) error {
	if state.LB.Type != "" {
		if config.AWS.Domain == "" {
			config.AWS.Domain = state.LB.Domain
		}

		if config.AWS.LBType == "" {
			config.AWS.LBType = state.LB.Type
		}
	}

	if lbExists(state.Stack.LBType) {
		return fmt.Errorf("bbl already has a %s load balancer attached, please remove the previous load balancer before attaching a new one", state.Stack.LBType)
	}

	if err := c.environmentValidator.Validate(state); err != nil {
		return err
	}

	certContents, err := ioutil.ReadFile(config.AWS.CertPath)
	if err != nil {
		return err
	}

	keyContents, err := ioutil.ReadFile(config.AWS.KeyPath)
	if err != nil {
		return err
	}

	state.LB.Cert = string(certContents)
	state.LB.Key = string(keyContents)

	if config.AWS.ChainPath != "" {
		chainContents, err := ioutil.ReadFile(config.AWS.ChainPath)
		if err != nil {
			return err
		}

		state.LB.Chain = string(chainContents)
	}

	if config.AWS.Domain != "" {
		state.LB.Domain = config.AWS.Domain
	}

	state.LB.Type = config.AWS.LBType

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
