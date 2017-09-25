package commands

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type azureClient interface {
	ValidateCredentials(subscriptionID, tenantID, clientID, clientSecret string) error
}

type AzureUp struct {
	azureClient        azureClient
	boshManager        boshManager
	cloudConfigManager cloudConfigManager
	envIDManager       envIDManager
	stateStore         stateStore
	terraformManager   terraformApplier
}

func NewAzureUp(azureClient azureClient,
	boshManager boshManager,
	cloudConfigManager cloudConfigManager,
	envIDManager envIDManager,
	stateStore stateStore,
	terraformManager terraformApplier) AzureUp {
	return AzureUp{
		azureClient:        azureClient,
		boshManager:        boshManager,
		cloudConfigManager: cloudConfigManager,
		envIDManager:       envIDManager,
		stateStore:         stateStore,
		terraformManager:   terraformManager,
	}
}

func (u AzureUp) Execute(config UpConfig, state storage.State) error {
	err := u.azureClient.ValidateCredentials(state.Azure.SubscriptionID, state.Azure.TenantID, state.Azure.ClientID, state.Azure.ClientSecret)
	if err != nil {
		return fmt.Errorf("Validate credentials: %s", err)
	}

	err = u.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	if config.NoDirector {
		if !state.BOSH.IsEmpty() {
			return errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
		}

		state.NoDirector = true
	}

	var opsFileContents []byte

	if config.OpsFile != "" {
		opsFileContents, err = ioutil.ReadFile(config.OpsFile)
		if err != nil {
			return fmt.Errorf("error reading ops-file contents: %v", err)
		}
	}

	state, err = u.envIDManager.Sync(state, config.Name)
	if err != nil {
		return err
	}

	err = u.stateStore.Set(state)
	if err != nil {
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
		state, err = u.boshManager.CreateJumpbox(state, terraformOutputs)
		if err != nil {
			return err
		}

		err = u.stateStore.Set(state)
		if err != nil {
			return err
		}

		state.BOSH.UserOpsFile = string(opsFileContents)

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
