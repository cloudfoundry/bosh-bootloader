package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

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
	stateStore         stateStore
	keyPairManager     keyPairManager
	gcpProvider        gcpProvider
	boshManager        boshManager
	cloudConfigManager cloudConfigManager
	logger             logger
	terraformManager   terraformApplier
	envIDManager       envIDManager
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

type gcpProvider interface {
	SetConfig(serviceAccountKey, projectID, region, zone string) error
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

type NewGCPUpArgs struct {
	StateStore         stateStore
	KeyPairManager     keyPairManager
	GCPProvider        gcpProvider
	TerraformManager   terraformApplier
	BoshManager        boshManager
	Logger             logger
	EnvIDManager       envIDManager
	CloudConfigManager cloudConfigManager
}

func NewGCPUp(args NewGCPUpArgs) GCPUp {
	return GCPUp{
		stateStore:         args.StateStore,
		keyPairManager:     args.KeyPairManager,
		gcpProvider:        args.GCPProvider,
		terraformManager:   args.TerraformManager,
		boshManager:        args.BoshManager,
		cloudConfigManager: args.CloudConfigManager,
		logger:             args.Logger,
		envIDManager:       args.EnvIDManager,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	state.IAAS = "gcp"
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

	gcpDetails, err := parseUpConfig(upConfig, state.GCP)
	if err != nil {
		return err
	}

	if err := fastFailConflictingGCPState(gcpDetails, state.GCP); err != nil {
		return err
	}

	state.GCP = gcpDetails

	if upConfig.NoDirector {
		if !state.BOSH.IsEmpty() {
			return errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
		}

		state.NoDirector = true
	}

	if err := u.validateState(state); err != nil {
		return err
	}

	if err := u.gcpProvider.SetConfig(state.GCP.ServiceAccountKey, state.GCP.ProjectID, state.GCP.Region, state.GCP.Zone); err != nil {
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

func parseUpConfig(upConfig GCPUpConfig, store storage.GCP) (storage.GCP, error) {
	var serviceAccountKey string
	if upConfig.ServiceAccountKey != "" {
		var err error
		serviceAccountKey, err = parseServiceAccountKey(upConfig.ServiceAccountKey)
		if err != nil {
			return storage.GCP{}, err
		}
	}

	gcpState := store
	if serviceAccountKey != "" {
		gcpState.ServiceAccountKey = serviceAccountKey
	}
	if upConfig.ProjectID != "" {
		gcpState.ProjectID = upConfig.ProjectID
	}
	if upConfig.Zone != "" {
		gcpState.Zone = upConfig.Zone
	}
	if upConfig.Region != "" {
		gcpState.Region = upConfig.Region
	}

	return gcpState, nil
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

func parseServiceAccountKey(serviceAccountKey string) (string, error) {
	var key string

	if _, err := os.Stat(serviceAccountKey); err != nil {
		key = serviceAccountKey
	} else {
		rawServiceAccountKey, err := ioutil.ReadFile(serviceAccountKey)
		if err != nil {
			return "", fmt.Errorf("error reading service account key from file: %v", err)
		}

		key = string(rawServiceAccountKey)
	}

	var tmp interface{}
	err := json.Unmarshal([]byte(key), &tmp)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling service account key (must be valid json): %v", err)
	}

	return key, err
}
