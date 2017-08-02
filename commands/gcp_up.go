package commands

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	yaml "gopkg.in/yaml.v2"

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
	keyPairManager               keyPairManager
	boshManager                  boshManager
	cloudConfigManager           cloudConfigManager
	logger                       logger
	terraformManager             terraformApplier
	envIDManager                 envIDManager
	gcpAvailabilityZoneRetriever gcpAvailabilityZoneRetriever
}

type GCPUpConfig struct {
	ServiceAccountKey string
	ProjectID         string
	Zone              string
	Region            string
	OpsFilePath       string
	Name              string
	NoDirector        bool
	Jumpbox           bool
}

type gcpKeyPairCreator interface {
	Create() (string, string, error)
}

type keyPairUpdater interface {
	Update() (storage.KeyPair, error)
}

type terraformManagerError interface {
	Error() string
	BBLState() (storage.State, error)
}

type boshManager interface {
	CreateDirector(bblState storage.State, terraformOutputs map[string]interface{}) (storage.State, error)
	CreateJumpbox(bblState storage.State, terraformOutputs map[string]interface{}) (storage.State, error)
	Delete(bblState storage.State, terraformOutputs map[string]interface{}) error
	DeleteJumpbox(bblState storage.State, terraformOutputs map[string]interface{}) error
	GetDeploymentVars(bblState storage.State, terraformOutputs map[string]interface{}) (string, error)
	Version() (string, error)
}

type envIDManager interface {
	Sync(storage.State, string) (storage.State, error)
}

type gcpAvailabilityZoneRetriever interface {
	GetZones(string) ([]string, error)
}

type NewGCPUpArgs struct {
	StateStore                   stateStore
	KeyPairManager               keyPairManager
	TerraformManager             terraformApplier
	BoshManager                  boshManager
	Logger                       logger
	EnvIDManager                 envIDManager
	CloudConfigManager           cloudConfigManager
	GCPAvailabilityZoneRetriever gcpAvailabilityZoneRetriever
}

func NewGCPUp(args NewGCPUpArgs) GCPUp {
	return GCPUp{
		stateStore:                   args.StateStore,
		keyPairManager:               args.KeyPairManager,
		terraformManager:             args.TerraformManager,
		boshManager:                  args.BoshManager,
		cloudConfigManager:           args.CloudConfigManager,
		logger:                       args.Logger,
		envIDManager:                 args.EnvIDManager,
		gcpAvailabilityZoneRetriever: args.GCPAvailabilityZoneRetriever,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	state.Jumpbox.Enabled = upConfig.Jumpbox

	err := u.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	var opsFileContents []byte
	if upConfig.OpsFilePath != "" {
		opsFileContents, err = ioutil.ReadFile(upConfig.OpsFilePath)
		if err != nil {
			return fmt.Errorf("error reading ops-file contents: %v", err)
		}
	}

	if upConfig.NoDirector {
		if !state.BOSH.IsEmpty() {
			return errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
		}

		state.NoDirector = true
	}

	if err := u.validateState(state); err != nil {
		return err
	}

	state, err = u.envIDManager.Sync(state, upConfig.Name)
	if err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	state, err = u.keyPairManager.Sync(state)
	if err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	state.GCP.Zones, err = u.gcpAvailabilityZoneRetriever.GetZones(state.GCP.Region)
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
		state.BOSH.UserOpsFile = string(opsFileContents)

		if upConfig.Jumpbox {
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

		err := u.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u GCPUp) validateState(state storage.State) error {
	switch {
	case state.GCP.ServiceAccountKey == "":
		return errors.New("GCP service account key must be provided")
	case state.GCP.ProjectID == "":
		return errors.New("GCP project ID must be provided")
	case state.GCP.Region == "":
		return errors.New("GCP region must be provided")
	case state.GCP.Zone == "":
		return errors.New("GCP zone must be provided")
	}

	return nil
}

func fastFailConflictingGCPState(configGCP storage.GCP, stateGCP storage.GCP) error {
	if stateGCP.Region != "" && stateGCP.Region != configGCP.Region {
		return errors.New(fmt.Sprintf("The region cannot be changed for an existing environment. The current region is %s.", stateGCP.Region))
	}

	if stateGCP.Zone != "" && stateGCP.Zone != configGCP.Zone {
		return errors.New(fmt.Sprintf("The zone cannot be changed for an existing environment. The current zone is %s.", stateGCP.Zone))
	}

	if stateGCP.ProjectID != "" && stateGCP.ProjectID != configGCP.ProjectID {
		return errors.New(fmt.Sprintf("The project id cannot be changed for an existing environment. The current project id is %s.", stateGCP.ProjectID))
	}

	return nil
}
