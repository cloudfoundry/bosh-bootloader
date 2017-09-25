package commands

import (
	"errors"
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var (
	marshal = yaml.Marshal
)

const (
	DIRECTOR_USERNAME = "admin"
)

type GCPUp struct {
	stateStore                   stateStore
	boshManager                  boshManager
	cloudConfigManager           cloudConfigManager
	terraformManager             terraformApplier
	envIDManager                 envIDManager
	gcpAvailabilityZoneRetriever gcpAvailabilityZoneRetriever
}

type gcpAvailabilityZoneRetriever interface {
	GetZones(string) ([]string, error)
}

func NewGCPUp(stateStore stateStore, terraformManager terraformApplier, boshManager boshManager,
	cloudConfigManager cloudConfigManager, envIDManager envIDManager, gcpAvailabilityZoneRetriever gcpAvailabilityZoneRetriever) GCPUp {
	return GCPUp{
		stateStore:                   stateStore,
		terraformManager:             terraformManager,
		boshManager:                  boshManager,
		cloudConfigManager:           cloudConfigManager,
		envIDManager:                 envIDManager,
		gcpAvailabilityZoneRetriever: gcpAvailabilityZoneRetriever,
	}
}

func (u GCPUp) Execute(config UpConfig, state storage.State) error {
	var err error
	state.GCP.Zones, err = u.gcpAvailabilityZoneRetriever.GetZones(state.GCP.Region)
	if err != nil {
		return fmt.Errorf("Retrieving availability zones: %s", err)
	}

	if err := u.stateStore.Set(state); err != nil {
		return fmt.Errorf("Save state after retrieving azs: %s", err)
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
