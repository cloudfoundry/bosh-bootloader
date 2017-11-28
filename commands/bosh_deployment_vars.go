package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type BOSHDeploymentVars struct {
	logger         logger
	boshManager    boshManager
	stateValidator stateValidator
	terraform      terraformManager
}

func NewBOSHDeploymentVars(logger logger, boshManager boshManager, stateValidator stateValidator, terraform terraformManager) BOSHDeploymentVars {
	return BOSHDeploymentVars{
		logger:         logger,
		boshManager:    boshManager,
		stateValidator: stateValidator,
		terraform:      terraform,
	}
}

func (b BOSHDeploymentVars) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := b.stateValidator.Validate()
	if err != nil {
		return err
	}

	if !state.NoDirector {
		err := fastFailBOSHVersion(b.boshManager)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b BOSHDeploymentVars) Execute(args []string, state storage.State) error {
	b.logger.Println(`Deprecation warning: the bosh-deployment-vars command has been deprecated and will be removed in bbl v6.0.0. The bosh deployment vars are stored in the vars directory.`)

	terraformOutputs, err := b.terraform.GetOutputs()
	if err != nil {
		return fmt.Errorf("get terraform outputs: %s", err)
	}

	vars := b.boshManager.GetDirectorDeploymentVars(state, terraformOutputs)
	b.logger.Println(vars)
	return nil
}
