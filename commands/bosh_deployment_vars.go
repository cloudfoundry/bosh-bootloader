package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

const (
	BOSHDeploymentVarsCommand = "bosh-deployment-vars"
)

type BOSHDeploymentVars struct {
	logger      logger
	boshManager boshManager
}

func NewBOSHDeploymentVars(logger logger, boshManager boshManager) BOSHDeploymentVars {
	return BOSHDeploymentVars{
		logger:      logger,
		boshManager: boshManager,
	}
}

func (b BOSHDeploymentVars) Execute(args []string, state storage.State) error {
	vars, err := b.boshManager.GetDeploymentVars(state)
	if err != nil {
		return err
	}
	b.logger.Println(vars)
	return nil
}
