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
	AWS AWSCreateLBsConfig
	GCP GCPCreateLBsConfig
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

func (c CreateLBs) CheckFastFails(subcommandFlags []string, state storage.State) error {
	config, err := parseFlags(subcommandFlags, state.IAAS, state.LB.Type)
	if err != nil {
		return err
	}

	if err := c.stateValidator.Validate(); err != nil {
		return fmt.Errorf("Validate state: %s", err)
	}

	if !lbExists(getLBType(config)) {
		return errors.New("--type is required")
	}

	if !(state.IAAS == "gcp" && getLBType(config) == "concourse") {
		err = c.certificateValidator.Validate("create-lbs", getCertPath(config), getKeyPath(config), getChainPath(config))
		if err != nil {
			return fmt.Errorf("Validate certificate: %s", err)
		}
	}

	if getLBType(config) == "concourse" && getDomain(config) != "" {
		return errors.New("--domain is not implemented for concourse load balancers. Remove the --domain flag and try again.")
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
	switch iaas {
	case "aws":
		lbFlags.String(&config.AWS.LBType, "type", existingLBType)
		lbFlags.String(&config.AWS.CertPath, "cert", "")
		lbFlags.String(&config.AWS.KeyPath, "key", "")
		lbFlags.String(&config.AWS.ChainPath, "chain", "")
		lbFlags.String(&config.AWS.Domain, "domain", "")
	case "gcp":
		lbFlags.String(&config.GCP.LBType, "type", existingLBType)
		lbFlags.String(&config.GCP.CertPath, "cert", "")
		lbFlags.String(&config.GCP.KeyPath, "key", "")
		lbFlags.String(&config.GCP.Domain, "domain", "")
	}

	if err := lbFlags.Parse(subcommandFlags); err != nil {
		return config, err
	}

	return config, nil
}

func getLBType(config CreateLBsConfig) string {
	if config.AWS.LBType != "" {
		return config.AWS.LBType
	}
	if config.GCP.LBType != "" {
		return config.GCP.LBType
	}
	return ""
}

func getCertPath(config CreateLBsConfig) string {
	if config.AWS.CertPath != "" {
		return config.AWS.CertPath
	}
	if config.GCP.CertPath != "" {
		return config.GCP.CertPath
	}
	return ""
}

func getKeyPath(config CreateLBsConfig) string {
	if config.AWS.KeyPath != "" {
		return config.AWS.KeyPath
	}
	if config.GCP.KeyPath != "" {
		return config.GCP.KeyPath
	}
	return ""
}

func getChainPath(config CreateLBsConfig) string {
	if config.AWS.ChainPath != "" {
		return config.AWS.ChainPath
	}
	return ""
}

func getDomain(config CreateLBsConfig) string {
	if config.AWS.Domain != "" {
		return config.AWS.Domain
	}
	if config.GCP.Domain != "" {
		return config.GCP.Domain
	}
	return ""
}
