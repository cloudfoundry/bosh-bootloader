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

type environmentValidator interface {
	Validate(state storage.State) error
}

func NewAWSCreateLBs(cloudConfigManager cloudConfigManager, stateStore stateStore,
	terraformManager terraformApplier, environmentValidator environmentValidator) AWSCreateLBs {
	return AWSCreateLBs{
		cloudConfigManager:   cloudConfigManager,
		stateStore:           stateStore,
		terraformManager:     terraformManager,
		environmentValidator: environmentValidator,
	}
}

func (c AWSCreateLBs) Execute(config AWSCreateLBsConfig, state storage.State) error {
	if lbExists(state.Stack.LBType) {
		return fmt.Errorf("bbl already has a %s load balancer attached, please remove the previous load balancer before attaching a new one", state.Stack.LBType)
	}

	if err := c.environmentValidator.Validate(state); err != nil {
		return err
	}

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
