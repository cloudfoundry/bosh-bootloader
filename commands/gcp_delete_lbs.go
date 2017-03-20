package commands

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type GCPDeleteLBs struct {
	cloudConfigManager cloudConfigManager
	logger             logger
	stateStore         stateStore
	terraformExecutor  terraformExecutor
}

func NewGCPDeleteLBs(logger logger, stateStore stateStore,
	terraformExecutor terraformExecutor, cloudConfigManager cloudConfigManager) GCPDeleteLBs {
	return GCPDeleteLBs{
		logger:             logger,
		stateStore:         stateStore,
		terraformExecutor:  terraformExecutor,
		cloudConfigManager: cloudConfigManager,
	}
}

func (g GCPDeleteLBs) Execute(state storage.State) error {
	err := fastFailTerraformVersion(g.terraformExecutor)
	if err != nil {
		return err
	}
	state.LB.Type = ""

	if !state.NoDirector {
		err = g.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	template := strings.Join([]string{terraform.VarsTemplate, terraformBOSHDirectorTemplate}, "\n")

	g.logger.Step("generating terraform template")
	tfState, err := g.terraformExecutor.Apply(state.GCP.ServiceAccountKey, state.EnvID, state.GCP.ProjectID,
		state.GCP.Zone, state.GCP.Region, "", "", "", template, state.TFState)

	switch err.(type) {
	case terraform.ExecutorApplyError:
		taErr := err.(terraform.ExecutorApplyError)
		state.TFState = taErr.TFState()
		if setErr := g.stateStore.Set(state); setErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(setErr)
			return errorList
		}
		return err
	case error:
		return err
	}
	g.logger.Step("finished applying terraform template")

	state.TFState = tfState

	err = g.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
