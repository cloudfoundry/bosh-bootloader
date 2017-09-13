package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type DeleteLBs struct {
	deleteLBs      DeleteLBsCmd
	logger         logger
	stateValidator stateValidator
	boshManager    boshManager
}

type DeleteLBsCmd interface {
	Execute(state storage.State) error
}

func NewDeleteLBs(deleteLBs DeleteLBsCmd,
	logger logger, stateValidator stateValidator, boshManager boshManager) DeleteLBs {
	return DeleteLBs{
		deleteLBs:      deleteLBs,
		logger:         logger,
		stateValidator: stateValidator,
		boshManager:    boshManager,
	}
}

func (d DeleteLBs) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := d.stateValidator.Validate()
	if err != nil {
		return err
	}

	if !state.NoDirector {
		err = fastFailBOSHVersion(d.boshManager)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d DeleteLBs) Execute(subcommandFlags []string, state storage.State) error {
	config, err := d.parseFlags(subcommandFlags)
	if err != nil {
		return err
	}

	if config.skipIfMissing && !lbExists(state.Stack.LBType) && !lbExists(state.LB.Type) {
		d.logger.Println("no lb type exists, skipping...")
		return nil
	}

	return d.deleteLBs.Execute(state)

}

func (DeleteLBs) parseFlags(subcommandFlags []string) (deleteLBsConfig, error) {
	lbFlags := flags.New("delete-lbs")

	config := deleteLBsConfig{}
	lbFlags.Bool(&config.skipIfMissing, "skip-if-missing", "", false)

	err := lbFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}
