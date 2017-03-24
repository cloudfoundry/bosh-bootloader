package commands

import (
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"github.com/cloudfoundry/multierror"
)

type GCPCreateLBs struct {
	terraformManager   terraformManager
	boshClientProvider boshClientProvider
	cloudConfigManager cloudConfigManager
	stateStore         stateStore
	logger             logger
}

type GCPCreateLBsConfig struct {
	LBType       string
	CertPath     string
	KeyPath      string
	Domain       string
	SkipIfExists bool
}

func NewGCPCreateLBs(terraformManager terraformManager,
	boshClientProvider boshClientProvider, cloudConfigManager cloudConfigManager,
	stateStore stateStore, logger logger) GCPCreateLBs {
	return GCPCreateLBs{
		terraformManager:   terraformManager,
		boshClientProvider: boshClientProvider,
		cloudConfigManager: cloudConfigManager,
		stateStore:         stateStore,
		logger:             logger,
	}
}

func (c GCPCreateLBs) Execute(config GCPCreateLBsConfig, state storage.State) error {
	err := c.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	err = c.checkFastFails(config, state)
	if err != nil {
		return err
	}

	if !state.NoDirector {
		boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername,
			state.BOSH.DirectorPassword)

		_, err := boshClient.Info()
		if err != nil {
			return BBLNotFound
		}
	}

	if config.SkipIfExists && config.LBType == state.LB.Type {
		c.logger.Step(fmt.Sprintf("lb type %q exists, skipping...", config.LBType))
		return nil
	}

	state.LB.Type = config.LBType

	var cert, key []byte
	if config.LBType == "cf" {
		state.LB.Domain = config.Domain

		cert, err = ioutil.ReadFile(config.CertPath)
		if err != nil {
			return err
		}

		state.LB.Cert = string(cert)

		key, err = ioutil.ReadFile(config.KeyPath)
		if err != nil {
			return err
		}

		state.LB.Key = string(key)
	}

	state, err = c.terraformManager.Apply(state)
	switch err.(type) {
	case terraform.ManagerApplyError:
		taError := err.(terraform.ManagerApplyError)
		state = taError.BBLState()
		if setErr := c.stateStore.Set(state); setErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(setErr)
			return errorList
		}
		return taError
	case error:
		return err
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

func (GCPCreateLBs) checkFastFails(config GCPCreateLBsConfig, state storage.State) error {
	if config.LBType == "" {
		return fmt.Errorf("--type is a required flag")
	}

	if config.LBType != "concourse" && config.LBType != "cf" {
		return fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse, cf", config.LBType)
	}

	if config.LBType == "cf" {
		errs := multierror.NewMultiError("create-lbs")
		if err := validateCertOrKeyFlag("cert", config.CertPath); err != nil {
			errs.Add(err)
		}
		if err := validateCertOrKeyFlag("key", config.KeyPath); err != nil {
			errs.Add(err)
		}

		if errs.Length() > 0 {
			return errs
		}
	}

	if state.IAAS != "gcp" {
		return fmt.Errorf("iaas type must be gcp")
	}

	return nil
}

func validateCertOrKeyFlag(flagName, path string) error {
	if path == "" {
		return fmt.Errorf("--%s is required", flagName)
	} else {
		body, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if string(body) == "" {
			return fmt.Errorf("provided %s file is empty", flagName)
		}
	}
	return nil
}
