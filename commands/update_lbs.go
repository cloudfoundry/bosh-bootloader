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
	domain        string
	skipIfMissing bool
}

type UpdateLBs struct {
	awsUpdateLBs         awsUpdateLBs
	gcpUpdateLBs         gcpUpdateLBs
	certificateValidator certificateValidator
	stateValidator       stateValidator
	logger               logger
	boshManager          boshManager
}

type awsUpdateLBs interface {
	Execute(AWSCreateLBsConfig, storage.State) error
}

type gcpUpdateLBs interface {
	Execute(GCPCreateLBsConfig, storage.State) error
}

func NewUpdateLBs(awsUpdateLBs awsUpdateLBs, gcpUpdateLBs gcpUpdateLBs, certificateValidator certificateValidator,
	stateValidator stateValidator, logger logger, boshManager boshManager) UpdateLBs {

	return UpdateLBs{
		awsUpdateLBs:         awsUpdateLBs,
		gcpUpdateLBs:         gcpUpdateLBs,
		certificateValidator: certificateValidator,
		stateValidator:       stateValidator,
		logger:               logger,
		boshManager:          boshManager,
	}
}

func (u UpdateLBs) Execute(subcommandFlags []string, state storage.State) error {
	config, err := u.parseFlags(subcommandFlags)
	if err != nil {
		return err
	}

	err = u.stateValidator.Validate()
	if err != nil {
		return err
	}

	if !state.NoDirector {
		err = fastFailBOSHVersion(u.boshManager)
		if err != nil {
			return err
		}
	}

	lbExists := lbExists(state.Stack.LBType) || lbExists(state.LB.Type)
	if config.skipIfMissing && !lbExists {
		u.logger.Println("no lb type exists, skipping...")
		return nil
	}

	if !lbExists {
		return LBNotFound
	}

	err = u.certificateValidator.Validate(UpdateLBsCommand, config.certPath, config.keyPath, config.chainPath)
	if err != nil {
		return err
	}

	switch state.IAAS {
	case "gcp":
		if err := u.gcpUpdateLBs.Execute(GCPCreateLBsConfig{
			LBType:   state.LB.Type,
			CertPath: config.certPath,
			KeyPath:  config.keyPath,
			Domain:   config.domain,
		}, state); err != nil {
			return err
		}
	case "aws":
		if err := u.awsUpdateLBs.Execute(AWSCreateLBsConfig{
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

func (u UpdateLBs) CheckFastFails(subcommandFlags []string, state storage.State) error {
	return nil
}

func (UpdateLBs) parseFlags(subcommandFlags []string) (updateLBConfig, error) {
	lbFlags := flags.New("update-lbs")

	config := updateLBConfig{}
	lbFlags.String(&config.certPath, "cert", "")
	lbFlags.String(&config.keyPath, "key", "")
	lbFlags.String(&config.chainPath, "chain", "")
	lbFlags.String(&config.domain, "domain", "")
	lbFlags.Bool(&config.skipIfMissing, "skip-if-missing", "", false)

	err := lbFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}
