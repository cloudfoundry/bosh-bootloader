package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type GCPDeleteLBs struct {
	cloudConfigManager cloudConfigManager
	stateStore         stateStore
	terraformManager   terraformManager
}

func NewGCPDeleteLBs(stateStore stateStore,
	terraformManager terraformManager, cloudConfigManager cloudConfigManager) GCPDeleteLBs {
	return GCPDeleteLBs{
		stateStore:         stateStore,
		terraformManager:   terraformManager,
		cloudConfigManager: cloudConfigManager,
	}
}

func (g GCPDeleteLBs) Execute(state storage.State) error {
	err := g.terraformManager.ValidateVersion()
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

	state, err = g.terraformManager.Apply(state)
	switch err.(type) {
	case terraform.ManagerApplyError:
		taErr := err.(terraform.ManagerApplyError)
		state = taErr.BBLState()
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

	err = g.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
