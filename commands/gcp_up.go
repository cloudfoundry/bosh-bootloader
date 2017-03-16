package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

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
	terraformExecutor  terraformExecutor
	zones              zones
	envIDManager       envIDManager
}

type GCPUpConfig struct {
	ServiceAccountKeyPath string
	ProjectID             string
	Zone                  string
	Region                string
	OpsFilePath           string
	Name                  string
	NoDirector            bool
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
}

type terraformExecutor interface {
	Apply(credentials, envID, projectID, zone, region, certPath, keyPath, domain, template, tfState string) (string, error)
	Destroy(serviceAccountKey, envID, projectID, zone, region, template, tfState string) (string, error)
	Version() (string, error)
}

type zones interface {
	Get(region string) []string
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

func NewGCPUp(stateStore stateStore, keyPairUpdater keyPairUpdater, gcpProvider gcpProvider, terraformExecutor terraformExecutor,
	boshManager boshManager, logger logger, zones zones, envIDManager envIDManager, cloudConfigManager cloudConfigManager) GCPUp {
	return GCPUp{
		stateStore:         stateStore,
		keyPairUpdater:     keyPairUpdater,
		gcpProvider:        gcpProvider,
		terraformExecutor:  terraformExecutor,
		boshManager:        boshManager,
		cloudConfigManager: cloudConfigManager,
		logger:             logger,
		zones:              zones,
		envIDManager:       envIDManager,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	err := fastFailTerraformVersion(u.terraformExecutor)
	if err != nil {
		return err
	}

	var opsFileContents []byte
	if !upConfig.empty() {
		var gcpDetails storage.GCP
		var err error
		gcpDetails, opsFileContents, err = u.parseUpConfig(upConfig)
		if err != nil {
			return err
		}

		state.IAAS = "gcp"

		if err := u.fastFailConflictingGCPState(gcpDetails, state.GCP); err != nil {
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

	var template string
	zones := u.zones.Get(state.GCP.Region)
	switch state.LB.Type {
	case "concourse":
		template = strings.Join([]string{terraform.VarsTemplate, terraformBOSHDirectorTemplate, terraformConcourseLBTemplate}, "\n")
	case "cf":
		terraformCFLBBackendService := generateBackendServiceTerraform(len(zones))
		instanceGroups := generateInstanceGroups(zones)

		if state.LB.Domain != "" {
			template = strings.Join([]string{terraform.VarsTemplate, terraformBOSHDirectorTemplate, terraformCFLBTemplate, instanceGroups, terraformCFLBBackendService, terraformCFDNSTemplate}, "\n")
		} else {
			template = strings.Join([]string{terraform.VarsTemplate, terraformBOSHDirectorTemplate, terraformCFLBTemplate, instanceGroups, terraformCFLBBackendService}, "\n")
		}
	default:
		template = strings.Join([]string{terraform.VarsTemplate, terraformBOSHDirectorTemplate}, "\n")
	}

	tfState, err := u.terraformExecutor.Apply(state.GCP.ServiceAccountKey,
		state.EnvID, state.GCP.ProjectID, state.GCP.Zone, state.GCP.Region, state.LB.Cert, state.LB.Key, state.LB.Domain,
		template, state.TFState,
	)
	switch err.(type) {
	case terraform.TerraformApplyError:
		taErr := err.(terraform.TerraformApplyError)
		state.TFState = taErr.TFState()
		if setErr := u.stateStore.Set(state); setErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(setErr)
			return errorList
		}
		return err
	case error:
		return err
	}

	state.TFState = tfState
	if err := u.stateStore.Set(state); err != nil {
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

func (u GCPUp) parseUpConfig(upConfig GCPUpConfig) (storage.GCP, []byte, error) {
	if upConfig.ServiceAccountKeyPath == "" {
		return storage.GCP{}, []byte{}, errors.New("GCP service account key must be provided")
	}

	sak, err := ioutil.ReadFile(upConfig.ServiceAccountKeyPath)
	if err != nil {
		return storage.GCP{}, []byte{}, fmt.Errorf("error reading service account key: %v", err)
	}

	var tmp interface{}
	err = json.Unmarshal(sak, &tmp)
	if err != nil {
		return storage.GCP{}, []byte{}, fmt.Errorf("error parsing service account key: %v", err)
	}

	var opsFileContents []byte
	if upConfig.OpsFilePath != "" {
		opsFileContents, err = ioutil.ReadFile(upConfig.OpsFilePath)
		if err != nil {
			return storage.GCP{}, []byte{}, fmt.Errorf("error reading ops-file contents: %v", err)
		}
	}

	return storage.GCP{
		ServiceAccountKey: string(sak),
		ProjectID:         upConfig.ProjectID,
		Zone:              upConfig.Zone,
		Region:            upConfig.Region,
	}, opsFileContents, nil
}

func (c GCPUpConfig) empty() bool {
	return c.ServiceAccountKeyPath == "" && c.ProjectID == "" && c.Region == "" && c.Zone == ""
}

func (u GCPUp) fastFailConflictingGCPState(configGCP storage.GCP, stateGCP storage.GCP) error {
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
