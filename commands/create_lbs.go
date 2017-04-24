package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const CreateLBsCommand = "create-lbs"

type CreateLBs struct {
	awsCreateLBs   awsCreateLBs
	gcpCreateLBs   gcpCreateLBs
	stateValidator stateValidator
	boshManager    boshManager
}

type lbConfig struct {
	lbType       string
	certPath     string
	keyPath      string
	chainPath    string
	domain       string
	skipIfExists bool
}

type gcpCreateLBs interface {
	Execute(GCPCreateLBsConfig, storage.State) error
}

type awsCreateLBs interface {
	Execute(AWSCreateLBsConfig, storage.State) error
}

func NewCreateLBs(awsCreateLBs awsCreateLBs, gcpCreateLBs gcpCreateLBs, stateValidator stateValidator, boshManager boshManager) CreateLBs {
	return CreateLBs{
		awsCreateLBs:   awsCreateLBs,
		gcpCreateLBs:   gcpCreateLBs,
		stateValidator: stateValidator,
		boshManager:    boshManager,
	}
}

func (c CreateLBs) Execute(args []string, state storage.State) error {
	if err := c.stateValidator.Validate(); err != nil {
		return err
	}

	if !state.NoDirector {
		err := fastFailBOSHVersion(c.boshManager)
		if err != nil {
			return err
		}
	}

	config, err := c.parseFlags(args)
	if err != nil {
		return err
	}

	switch state.IAAS {
	case "gcp":
		if err := c.gcpCreateLBs.Execute(GCPCreateLBsConfig{
			LBType:       config.lbType,
			CertPath:     config.certPath,
			KeyPath:      config.keyPath,
			Domain:       config.domain,
			SkipIfExists: config.skipIfExists,
		}, state); err != nil {
			return err
		}
	case "aws":
		if err := c.awsCreateLBs.Execute(AWSCreateLBsConfig{
			LBType:       config.lbType,
			CertPath:     config.certPath,
			KeyPath:      config.keyPath,
			ChainPath:    config.chainPath,
			Domain:       config.domain,
			SkipIfExists: config.skipIfExists,
		}, state); err != nil {
			return err
		}
	}

	return nil
}

func (CreateLBs) parseFlags(subcommandFlags []string) (lbConfig, error) {
	lbFlags := flags.New("create-lbs")

	config := lbConfig{}
	lbFlags.String(&config.lbType, "type", "")
	lbFlags.String(&config.certPath, "cert", "")
	lbFlags.String(&config.keyPath, "key", "")
	lbFlags.String(&config.chainPath, "chain", "")
	lbFlags.String(&config.domain, "domain", "")
	lbFlags.Bool(&config.skipIfExists, "skip-if-exists", "", false)

	if err := lbFlags.Parse(subcommandFlags); err != nil {
		return config, err
	}

	return config, nil
}
