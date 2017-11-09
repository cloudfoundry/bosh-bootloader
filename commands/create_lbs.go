package commands

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type CreateLBs struct {
	createLBsCmd         CreateLBsCmd
	boshManager          boshManager
	certificateValidator certificateValidator
	logger               logger
	stateValidator       stateValidator
}

type CreateLBsCmd interface {
	Execute(createLBsConfig CreateLBsConfig, state storage.State) error
}

type CreateLBsConfig struct {
	LBType    string
	CertPath  string
	KeyPath   string
	ChainPath string
	Domain    string
}

var LBNotFound error = errors.New("no load balancer has been found for this bbl environment")

func NewCreateLBs(createLBsCmd CreateLBsCmd, logger logger, stateValidator stateValidator, certificateValidator certificateValidator, boshManager boshManager) CreateLBs {
	return CreateLBs{
		createLBsCmd:         createLBsCmd,
		boshManager:          boshManager,
		logger:               logger,
		stateValidator:       stateValidator,
		certificateValidator: certificateValidator,
	}
}

func (c CreateLBs) validateLBArgs(config CreateLBsConfig, iaas string) error {
	if !lbExists(config.LBType) {
		return errors.New("--type is required")
	}

	if !(iaas == "gcp" && config.LBType == "concourse") {
		err := c.certificateValidator.Validate("create-lbs", config.CertPath, config.KeyPath, config.ChainPath)
		if err != nil {
			return fmt.Errorf("Validate certificate: %s", err)
		}
	}

	if config.LBType == "concourse" && config.Domain != "" {
		return errors.New("--domain is not implemented for concourse load balancers. Remove the --domain flag and try again.")
	}

	return nil
}

func (c CreateLBs) CheckFastFails(subcommandFlags []string, state storage.State) error {
	config, err := parseFlags(subcommandFlags, state.IAAS, state.LB.Type)
	if err != nil {
		return err
	}

	if err := c.validateLBArgs(config, state.IAAS); err != nil {
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

	err = c.createLBsCmd.Execute(config, state)
	if err != nil {
		return err
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
