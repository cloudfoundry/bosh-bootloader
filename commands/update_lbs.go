package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const UpdateLBsCommand = "update-lbs"

type updateLBConfig struct {
	certPath      string
	keyPath       string
	chainPath     string
	systemDomain  string
	skipIfMissing bool
}

type UpdateLBs struct {
	awsUpdateLBs         awsUpdateLBs
	gcpUpdateLBs         gcpUpdateLBs
	certificateValidator certificateValidator
	stateValidator       stateValidator
	logger               logger
}

type awsUpdateLBs interface {
	Execute(AWSCreateLBsConfig, storage.State) error
}

type gcpUpdateLBs interface {
	Execute(GCPCreateLBsConfig, storage.State) error
}

func NewUpdateLBs(awsUpdateLBs awsUpdateLBs, gcpUpdateLBs gcpUpdateLBs, certificateValidator certificateValidator,
	stateValidator stateValidator, logger logger) UpdateLBs {

	return UpdateLBs{
		awsUpdateLBs:         awsUpdateLBs,
		gcpUpdateLBs:         gcpUpdateLBs,
		certificateValidator: certificateValidator,
		stateValidator:       stateValidator,
		logger:               logger,
	}
}

func (c UpdateLBs) Execute(subcommandFlags []string, state storage.State) error {
	config, err := c.parseFlags(subcommandFlags)
	if err != nil {
		return err
	}

	err = c.stateValidator.Validate()
	if err != nil {
		return err
	}

	lbExists := lbExists(state.Stack.LBType) || lbExists(state.LB.Type)
	if config.skipIfMissing && !lbExists {
		c.logger.Println("no lb type exists, skipping...")
		return nil
	}

	if !lbExists {
		return LBNotFound
	}

	err = c.certificateValidator.Validate(UpdateLBsCommand, config.certPath, config.keyPath, config.chainPath)
	if err != nil {
		return err
	}

	switch state.IAAS {
	case "gcp":
		if err := c.gcpUpdateLBs.Execute(GCPCreateLBsConfig{
			LBType:       state.LB.Type,
			CertPath:     config.certPath,
			KeyPath:      config.keyPath,
			SystemDomain: config.systemDomain,
		}, state); err != nil {
			return err
		}
	case "aws":
		if err := c.awsUpdateLBs.Execute(AWSCreateLBsConfig{
			LBType:    state.Stack.LBType,
			CertPath:  config.certPath,
			KeyPath:   config.keyPath,
			ChainPath: config.chainPath,
		}, state); err != nil {
			return err
		}
	}
	return nil
}

func (UpdateLBs) parseFlags(subcommandFlags []string) (updateLBConfig, error) {
	lbFlags := flags.New("update-lbs")

	config := updateLBConfig{}
	lbFlags.String(&config.certPath, "cert", "")
	lbFlags.String(&config.keyPath, "key", "")
	lbFlags.String(&config.chainPath, "chain", "")
	lbFlags.String(&config.systemDomain, "d", "")
	lbFlags.Bool(&config.skipIfMissing, "skip-if-missing", "", false)

	err := lbFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}
