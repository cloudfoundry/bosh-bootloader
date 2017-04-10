package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

var (
	marshal = yaml.Marshal
)

const (
	DIRECTOR_USERNAME = "admin"
)

type GCPUp struct {
	stateStore         stateStore
	keyPairUpdater     keyPairUpdater
	gcpProvider        gcpProvider
	boshManager        boshManager
	cloudConfigManager cloudConfigManager
	logger             logger
	terraformManager   terraformManager
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
}

type gcpKeyPairCreator interface {
	Create() (string, string, error)
}

type keyPairUpdater interface {
	Update() (storage.KeyPair, error)
}

type gcpProvider interface {
	SetConfig(serviceAccountKey, projectID, zone string) error
}

type terraformManager interface {
	Destroy(bblState storage.State) (storage.State, error)
	Apply(bblState storage.State) (storage.State, error)
	GetOutputs(tfState, lbType string, domainExists bool) (terraform.Outputs, error)
	Version() (string, error)
	ValidateVersion() error
}

type terraformManagerError interface {
	Error() string
	BBLState() (storage.State, error)
}

type boshManager interface {
	Create(storage.State, []byte) (storage.State, error)
	Delete(storage.State) error
	GetDeploymentVars(storage.State) (string, error)
	Version() (string, error)
}

type envIDManager interface {
	Sync(storage.State, string) (string, error)
}

func NewGCPUp(stateStore stateStore, keyPairUpdater keyPairUpdater, gcpProvider gcpProvider, terraformManager terraformManager,
	boshManager boshManager, logger logger, envIDManager envIDManager, cloudConfigManager cloudConfigManager) GCPUp {
	return GCPUp{
		stateStore:         stateStore,
		keyPairUpdater:     keyPairUpdater,
		gcpProvider:        gcpProvider,
		terraformManager:   terraformManager,
		boshManager:        boshManager,
		cloudConfigManager: cloudConfigManager,
		logger:             logger,
		envIDManager:       envIDManager,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	err := u.terraformManager.ValidateVersion()
	if err != nil {
		return err
	}

	var opsFileContents []byte
	if !upConfig.empty() {
		var gcpDetails storage.GCP
		var err error
		gcpDetails, opsFileContents, err = parseUpConfig(upConfig)
		if err != nil {
			return err
		}

		state.IAAS = "gcp"

		if err := fastFailConflictingGCPState(gcpDetails, state.GCP); err != nil {
			return err
		}

		if upConfig.NoDirector {
			if !state.BOSH.IsEmpty() {
				return errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
			}

			state.NoDirector = true
		}

		state.GCP = gcpDetails
	}

	if err := u.validateState(state); err != nil {
		return err
	}

	if err := u.gcpProvider.SetConfig(state.GCP.ServiceAccountKey, state.GCP.ProjectID, state.GCP.Zone); err != nil {
		return err
	}

	envID, err := u.envIDManager.Sync(state, upConfig.Name)
	if err != nil {
		return err
	}

	state.EnvID = envID

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	if state.KeyPair.IsEmpty() {
		keyPair, err := u.keyPairUpdater.Update()
		if err != nil {
			return err
		}
		state.KeyPair = keyPair
		if err := u.stateStore.Set(state); err != nil {
			return err
		}
	}

	state, err = u.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, u.stateStore)
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return err
	}

	if !state.NoDirector {
		state, err = u.boshManager.Create(state, opsFileContents)
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

func parseUpConfig(upConfig GCPUpConfig) (storage.GCP, []byte, error) {
	if upConfig.ServiceAccountKey == "" {
		return storage.GCP{}, []byte{}, errors.New("GCP service account key must be provided")
	}

	serviceAccountKey, err := parseServiceAccountKey(upConfig.ServiceAccountKey)
	if err != nil {
		return storage.GCP{}, []byte{}, err
	}

	var opsFileContents []byte
	if upConfig.OpsFilePath != "" {
		opsFileContents, err = ioutil.ReadFile(upConfig.OpsFilePath)
		if err != nil {
			return storage.GCP{}, []byte{}, fmt.Errorf("error reading ops-file contents: %v", err)
		}
	}

	return storage.GCP{
		ServiceAccountKey: serviceAccountKey,
		ProjectID:         upConfig.ProjectID,
		Zone:              upConfig.Zone,
		Region:            upConfig.Region,
	}, opsFileContents, nil
}

func (c GCPUpConfig) empty() bool {
	return c.ServiceAccountKey == "" && c.ProjectID == "" && c.Region == "" && c.Zone == ""
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
	var tmp interface{}
	rawServiceAccountKey, err := ioutil.ReadFile(serviceAccountKey)
	if err != nil {
		err = json.Unmarshal([]byte(serviceAccountKey), &tmp)
		if err != nil {
			return "", fmt.Errorf("error reading or parsing service account key (must be valid json or a file containing valid json): %v", err)
		}
		return serviceAccountKey, nil
	} else {
		err = json.Unmarshal(rawServiceAccountKey, &tmp)
		if err != nil {
			return "", fmt.Errorf("error reading or parsing service account key (must be valid json or a file containing valid json): %v", err)
		}
		return string(rawServiceAccountKey), nil
	}
}
