package commands

import (
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPCreateLBs struct {
	terraformManager          terraformApplier
	cloudConfigManager        cloudConfigManager
	stateStore                stateStore
	environmentValidator      EnvironmentValidator
	availabilityZoneRetriever availabilityZoneRetriever
}

type GCPCreateLBsConfig struct {
	LBType   string
	CertPath string
	KeyPath  string
	Domain   string
}

type availabilityZoneRetriever interface {
	GetZones(region string) ([]string, error)
}

func NewGCPCreateLBs(terraformManager terraformApplier,
	cloudConfigManager cloudConfigManager,
	stateStore stateStore, environmentValidator EnvironmentValidator,
	availabilityZoneRetriever availabilityZoneRetriever) GCPCreateLBs {
	return GCPCreateLBs{
		terraformManager:          terraformManager,
		cloudConfigManager:        cloudConfigManager,
		stateStore:                stateStore,
		environmentValidator:      environmentValidator,
		availabilityZoneRetriever: availabilityZoneRetriever,
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

	state.GCP.Zones, err = c.availabilityZoneRetriever.GetZones(state.GCP.Region)
	if err != nil {
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
