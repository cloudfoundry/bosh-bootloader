package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type DeleteLBs struct {
	logger               logger
	stateValidator       stateValidator
	boshManager          boshManager
	cloudConfigManager   cloudConfigManager
	stateStore           stateStore
	environmentValidator environmentValidator
	terraformManager     terraformApplier
}

type config struct {
	skipIfMissing bool
}

func NewDeleteLBs(logger logger, stateValidator stateValidator, boshManager boshManager,
	cloudConfigManager cloudConfigManager, stateStore stateStore,
	environmentValidator environmentValidator, terraformManager terraformApplier) DeleteLBs {
	return DeleteLBs{
		logger:               logger,
		stateValidator:       stateValidator,
		boshManager:          boshManager,
		cloudConfigManager:   cloudConfigManager,
		stateStore:           stateStore,
		environmentValidator: environmentValidator,
		terraformManager:     terraformManager,
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

	if config.skipIfMissing && !lbExists(state.LB.Type) {
		d.logger.Println("no lb type exists, skipping...")
		return nil
	}

	err = d.environmentValidator.Validate(state)
	if err != nil {
		return fmt.Errorf("Environment validate: %s", err)
	}

	if !lbExists(state.LB.Type) {
		return LBNotFound
	}

	state.LB = storage.LB{}

	if !state.NoDirector {
		err = d.cloudConfigManager.Update(state)
		if err != nil {
			return fmt.Errorf("Update cloud config: %s", err)
		}
	}

	state, err = d.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, d.stateStore)
	}

	err = d.stateStore.Set(state)
	if err != nil {
		return fmt.Errorf("Save state after delete lbs: %s", err)
	}

	return nil
}

func (DeleteLBs) parseFlags(subcommandFlags []string) (config, error) {
	lbFlags := flags.New("delete-lbs")

	c := config{}
	lbFlags.Bool(&c.skipIfMissing, "skip-if-missing", "", false)

	err := lbFlags.Parse(subcommandFlags)
	if err != nil {
		return config{}, err
	}

	return c, nil
}
