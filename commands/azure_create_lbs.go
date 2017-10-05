package commands

import (
	"io/ioutil"

	// yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AzureCreateLBs struct {
	terraformManager          terraformApplier
	cloudConfigManager        cloudConfigManager
	stateStore                stateStore
	environmentValidator      EnvironmentValidator
}

type AzureCreateLBsConfig struct {
	LBType   string
	CertPath string
	KeyPath  string
	Domain   string
}

func NewAzureCreateLBs(terraformManager terraformApplier,
	cloudConfigManager cloudConfigManager,
	stateStore stateStore, environmentValidator EnvironmentValidator) AzureCreateLBs {
	return AzureCreateLBs{
		terraformManager:          terraformManager,
		cloudConfigManager:        cloudConfigManager,
		stateStore:                stateStore,
		environmentValidator:      environmentValidator,
	}
}

func (c AzureCreateLBs) Execute(config CreateLBsConfig, state storage.State) error {
	if state.LB.Type != "" {
		if config.Azure.Domain == "" {
			config.Azure.Domain = state.LB.Domain
		}
	}

	err := c.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	if err := c.environmentValidator.Validate(state); err != nil {
		return err
	}

	state.LB.Type = config.Azure.LBType

	var cert, key []byte
	if config.Azure.LBType == "cf" {
		state.LB.Domain = config.Azure.Domain

		cert, err = ioutil.ReadFile(config.Azure.CertPath)
		if err != nil {
			return err
		}

		state.LB.Cert = string(cert)

		key, err = ioutil.ReadFile(config.Azure.KeyPath)
		if err != nil {
			return err
		}

		state.LB.Key = string(key)
	}

	state, err = c.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, c.stateStore)
	}

	if err := c.stateStore.Set(state); err != nil {
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