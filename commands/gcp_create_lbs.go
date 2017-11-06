package commands

import (
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
		if config.GCP.Domain == "" {
			config.GCP.Domain = state.LB.Domain
		}
	}

	err := c.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	if err := c.environmentValidator.Validate(state); err != nil {
		return err
	}

	state.LB.Type = config.GCP.LBType

	var cert, key []byte
	if config.GCP.LBType == "cf" {
		state.LB.Domain = config.GCP.Domain

		cert, err = ioutil.ReadFile(config.GCP.CertPath)
		if err != nil {
			return err
		}

		state.LB.Cert = string(cert)

		key, err = ioutil.ReadFile(config.GCP.KeyPath)
		if err != nil {
			return err
		}

		state.LB.Key = string(key)
	}

	if err := c.terraformManager.Init(state); err != nil {
		return err
	}

	state, err = c.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, state, c.stateStore)
	}

	if err := c.stateStore.Set(state); err != nil {
		return err
	}

	if !state.NoDirector {
		err = c.cloudConfigManager.Initialize(state)
		if err != nil {
			return err
		}
		err = c.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	return nil
}
