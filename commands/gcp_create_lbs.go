package commands

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var (
	marshal = yaml.Marshal
)

type GCPCreateLBs struct {
	terraformManager     terraformManager
	cloudConfigManager   cloudConfigManager
	stateStore           stateStore
	environmentValidator EnvironmentValidator
}

type GCPCreateLBsConfig struct {
	LBType   string
	CertPath string
	KeyPath  string
	Domain   string
}

func NewGCPCreateLBs(terraformManager terraformManager,
	cloudConfigManager cloudConfigManager,
	stateStore stateStore, environmentValidator EnvironmentValidator,
) GCPCreateLBs {
	return GCPCreateLBs{
		terraformManager:     terraformManager,
		cloudConfigManager:   cloudConfigManager,
		stateStore:           stateStore,
		environmentValidator: environmentValidator,
	}
}

func (c GCPCreateLBs) Execute(config CreateLBsConfig, state storage.State) error {
	if state.LB.Type != "" {
		if config.Domain == "" {
			config.Domain = state.LB.Domain
		}
	}

	err := c.terraformManager.ValidateVersion()
	if err != nil {
		return fmt.Errorf("validate terraform version: %s", err)
	}

	if err := c.environmentValidator.Validate(state); err != nil {
		return fmt.Errorf("validate environment: %s", err)
	}

	state.LB.Type = config.LBType

	var cert, key []byte
	if config.LBType == "cf" {
		state.LB.Domain = config.Domain

		cert, err = ioutil.ReadFile(config.CertPath)
		if err != nil {
			return fmt.Errorf("read cert: %s", err)
		}

		state.LB.Cert = string(cert)

		key, err = ioutil.ReadFile(config.KeyPath)
		if err != nil {
			return fmt.Errorf("read key: %s", err)
		}

		state.LB.Key = string(key)
	}

	if err := c.terraformManager.Init(state); err != nil {
		return fmt.Errorf("initialize terraform: %s", err)
	}

	state, err = c.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, state, c.stateStore)
	}

	if err := c.stateStore.Set(state); err != nil {
		return fmt.Errorf("save state: %s", err)
	}

	if !state.NoDirector {
		err = c.cloudConfigManager.Initialize(state)
		if err != nil {
			return fmt.Errorf("initialize cloud config: %s", err)
		}
		err = c.cloudConfigManager.Update(state)
		if err != nil {
			return fmt.Errorf("update cloud config: %s", err)
		}
	}

	return nil
}
