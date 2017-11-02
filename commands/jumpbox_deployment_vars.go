package commands

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type JumpboxDeploymentVars struct {
	logger         logger
	boshManager    boshManager
	stateValidator stateValidator
	terraform      terraformManager
}

func NewJumpboxDeploymentVars(logger logger, boshManager boshManager, stateValidator stateValidator, terraform terraformManager) JumpboxDeploymentVars {
	return JumpboxDeploymentVars{
		logger:         logger,
		boshManager:    boshManager,
		stateValidator: stateValidator,
		terraform:      terraform,
	}
}

func (b JumpboxDeploymentVars) CheckFastFails(subcommandFlags []string, state storage.State) error {
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

func (b JumpboxDeploymentVars) Execute(args []string, state storage.State) error {
	terraformOutputs, err := b.terraform.GetOutputs()
	if err != nil {
		return fmt.Errorf("get terraform outputs: %s", err)
	}

	vars := b.boshManager.GetJumpboxDeploymentVars(state, terraformOutputs)
	b.logger.Println(vars)
	return nil
}
