package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type CreateLBs struct {
	boshManager          boshManager
	logger               logger
	stateValidator       stateValidator
	lbArgsHandler        lbArgsHandler
	cloudConfigManager   cloudConfigManager
	terraformManager     terraformManager
	stateStore           stateStore
	environmentValidator environmentValidator
}

type CreateLBsConfig struct {
	LBType    string
	CertPath  string
	KeyPath   string
	ChainPath string
	Domain    string
}

var LBNotFound error = errors.New("no load balancer has been found for this bbl environment")

func NewCreateLBs(
	logger logger,
	stateValidator stateValidator,
	boshManager boshManager,
	lbArgsHandler lbArgsHandler,
	cloudConfigManager cloudConfigManager,
	terraformManager terraformManager,
	stateStore stateStore,
	environmentValidator environmentValidator,
) CreateLBs {
	return CreateLBs{
		boshManager:          boshManager,
		logger:               logger,
		stateValidator:       stateValidator,
		lbArgsHandler:        lbArgsHandler,
		cloudConfigManager:   cloudConfigManager,
		terraformManager:     terraformManager,
		stateStore:           stateStore,
		environmentValidator: environmentValidator,
	}
}

func (c CreateLBs) CheckFastFails(subcommandFlags []string, state storage.State) error {
	config, err := parseFlags(subcommandFlags, state.IAAS, state.LB.Type)
	if err != nil {
		return err
	}

	if _, err := c.lbArgsHandler.GetLBState(state.IAAS, config); err != nil {
		return err
	}

	if err := c.stateValidator.Validate(); err != nil {
		return fmt.Errorf("Validate state: %s", err)
	}

	if !state.NoDirector {
		err := fastFailBOSHVersion(c.boshManager)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c CreateLBs) Execute(args []string, state storage.State) error {
	config, err := parseFlags(args, state.IAAS, state.LB.Type)
	if err != nil {
		return err
	}

	err = c.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	newLBState, err := c.lbArgsHandler.GetLBState(state.IAAS, config)
	if err != nil {
		return err
	}
	state.LB = c.lbArgsHandler.Merge(newLBState, state.LB)

	if err := c.environmentValidator.Validate(state); err != nil {
		return err
	}

	if err := c.stateStore.Set(state); err != nil {
		return fmt.Errorf("saving state before terraform init: %s", err)
	}

	if err := c.terraformManager.Init(state); err != nil {
		return err
	}

	state, err = c.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, state, c.stateStore)
	}

	if err := c.stateStore.Set(state); err != nil {
		return fmt.Errorf("saving state after terraform apply: %s", err)
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

func parseFlags(subcommandFlags []string, iaas string, existingLBType string) (CreateLBsConfig, error) {
	lbFlags := flags.New("create-lbs")

	config := CreateLBsConfig{}
	lbFlags.String(&config.LBType, "type", existingLBType)
	lbFlags.String(&config.CertPath, "cert", "")
	lbFlags.String(&config.KeyPath, "key", "")
	lbFlags.String(&config.Domain, "domain", "")

	if iaas == "aws" {
		lbFlags.String(&config.ChainPath, "chain", "")
	}

	if err := lbFlags.Parse(subcommandFlags); err != nil {
		return config, err
	}

	return config, nil
}
