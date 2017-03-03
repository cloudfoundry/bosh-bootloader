package commands

import "github.com/cloudfoundry/bosh-bootloader/storage"

const (
	BOSHDeploymentVarsCommand = "bosh-deployment-vars"
)

type BOSHDeploymentVars struct {
	logger            logger
	boshManager       boshManager
	terraformExecutor terraformExecutor
}

func NewBOSHDeploymentVars(logger logger, boshManager boshManager, terraformExecutor terraformExecutor) BOSHDeploymentVars {
	return BOSHDeploymentVars{
		logger:            logger,
		boshManager:       boshManager,
		terraformExecutor: terraformExecutor,
	}
}

func (b BOSHDeploymentVars) Execute(args []string, state storage.State) error {
	err := fastFailTerraformVersion(b.terraformExecutor)
	if err != nil {
		return err
	}
	vars, err := b.boshManager.GetDeploymentVars(state)
	if err != nil {
		return err
	}
	b.logger.Println(vars)
	return nil
}
