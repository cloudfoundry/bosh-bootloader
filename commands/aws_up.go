package commands

import (
	"errors"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSUp struct {
	boshManager        boshManager
	cloudConfigManager cloudConfigManager
	stateStore         stateStore
	envIDManager       envIDManager
	terraformManager   terraformApplier
}

func NewAWSUp(boshManager boshManager, cloudConfigManager cloudConfigManager,
	stateStore stateStore, envIDManager envIDManager, terraformManager terraformApplier) AWSUp {
	return AWSUp{
		boshManager:        boshManager,
		cloudConfigManager: cloudConfigManager,
		stateStore:         stateStore,
		envIDManager:       envIDManager,
		terraformManager:   terraformManager,
	}
}

func (u AWSUp) Execute(config UpConfig, state storage.State) error {
	state.IAAS = "aws"

	if config.NoDirector {
		if !state.BOSH.IsEmpty() {
			return errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
		}

		state.NoDirector = true
	}

	var err error
	state, err = u.envIDManager.Sync(state, config.Name)
	if err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	state, err = u.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, u.stateStore)
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return err
	}

	terraformOutputs, err := u.terraformManager.GetOutputs(state)
	if err != nil {
		return err
	}

	if !state.NoDirector {
		opsFile := []byte{}
		if config.OpsFile != "" {
			opsFile, err = ioutil.ReadFile(config.OpsFile)
			if err != nil {
				return err
			}
		}
		state.BOSH.UserOpsFile = string(opsFile)

		if config.Jumpbox {
			state.Jumpbox.Enabled = true
			state, err = u.boshManager.CreateJumpbox(state, terraformOutputs)
			if err != nil {
				return err
			}
		}

		state, err = u.boshManager.CreateDirector(state, terraformOutputs)
		switch err.(type) {
		case bosh.ManagerCreateError:
			bcErr := err.(bosh.ManagerCreateError)
			if setErr := u.stateStore.Set(bcErr.State()); setErr != nil {
				errorList := helpers.Errors{}
				errorList.Add(err)
				errorList.Add(setErr)
				return errorList
			}
			return err
		case error:
			return err
		}

		err = u.stateStore.Set(state)
		if err != nil {
			return err
		}

		err = u.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}
	return nil
}
