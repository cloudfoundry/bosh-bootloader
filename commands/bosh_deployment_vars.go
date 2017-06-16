package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

const (
	BOSHDeploymentVarsCommand = "bosh-deployment-vars"
)

type BOSHDeploymentVars struct {
	logger         logger
	boshManager    boshManager
	stateValidator stateValidator
}

func NewBOSHDeploymentVars(logger logger, boshManager boshManager, stateValidator stateValidator) BOSHDeploymentVars {
	return BOSHDeploymentVars{
		logger:         logger,
		boshManager:    boshManager,
		stateValidator: stateValidator,
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
	vars, err := b.boshManager.GetDeploymentVars(state)
	if err != nil {
		return err
	}
	b.logger.Println(vars)
	return nil
}
