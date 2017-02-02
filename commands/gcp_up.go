package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/boshinit"
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
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
	stateStore           stateStore
	keyPairUpdater       keyPairUpdater
	gcpProvider          gcpProvider
	boshDeployer         boshDeployer
	logger               logger
	boshClientProvider   boshClientProvider
	cloudConfigGenerator gcpCloudConfigGenerator
	terraformOutputter   terraformOutputter
	terraformExecutor    terraformExecutor
	zones                zones
}

type GCPUpConfig struct {
	ServiceAccountKeyPath string
	ProjectID             string
	Zone                  string
	Region                string
}

type directorOutputs struct {
	directorPassword       string
	directorSSLCA          string
	directorSSLCertificate string
	directorSSLPrivateKey  string
}

type gcpCloudConfigGenerator interface {
	Generate(gcp.CloudConfigInput) (gcp.CloudConfig, error)
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

type terraformExecutor interface {
	Apply(credentials, envID, projectID, zone, region, certPath, keyPath, domain, template, tfState string) (string, error)
	Destroy(serviceAccountKey, envID, projectID, zone, region, template, tfState string) (string, error)
}

type terraformOutputter interface {
	Get(tfState, outputName string) (string, error)
}

type zones interface {
	Get(region string) []string
}

type boshDeployer interface {
	Deploy(bosh.DeployInput) (bosh.DeployOutput, error)
}

func NewGCPUp(stateStore stateStore, keyPairUpdater keyPairUpdater, gcpProvider gcpProvider, terraformExecutor terraformExecutor, boshDeployer boshDeployer,
	logger logger, boshClientProvider boshClientProvider, cloudConfigGenerator gcpCloudConfigGenerator,
	terraformOutputter terraformOutputter, zones zones) GCPUp {
	return GCPUp{
		stateStore:           stateStore,
		keyPairUpdater:       keyPairUpdater,
		gcpProvider:          gcpProvider,
		terraformExecutor:    terraformExecutor,
		boshDeployer:         boshDeployer,
		logger:               logger,
		boshClientProvider:   boshClientProvider,
		cloudConfigGenerator: cloudConfigGenerator,
		terraformOutputter:   terraformOutputter,
		zones:                zones,
	}
}

func (u GCPUp) Execute(upConfig GCPUpConfig, state storage.State) error {
	if !upConfig.empty() {
		gcpDetails, err := u.parseUpConfig(upConfig)
		if err != nil {
			return err
		}

		state.IAAS = "gcp"

		if err := u.fastFailConflictingGCPState(gcpDetails, state.GCP); err != nil {
			return err
		}

		state.GCP = gcpDetails
	}

	if err := u.validateState(state); err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	if err := u.gcpProvider.SetConfig(state.GCP.ServiceAccountKey, state.GCP.ProjectID, state.GCP.Zone); err != nil {
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
		template = strings.Join([]string{terraformVarsTemplate, terraformBOSHDirectorTemplate, terraformConcourseLBTemplate}, "\n")
	case "cf":
		terraformCFLBBackendService := generateBackendServiceTerraform(len(zones))
		instanceGroups := generateInstanceGroups(zones)
		template = strings.Join([]string{terraformVarsTemplate, terraformBOSHDirectorTemplate, terraformCFLBTemplate, instanceGroups, terraformCFLBBackendService}, "\n")
	default:
		template = strings.Join([]string{terraformVarsTemplate, terraformBOSHDirectorTemplate}, "\n")
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

	externalIP, err := u.terraformOutputter.Get(state.TFState, "external_ip")
	if err != nil {
		return err
	}

	networkName, err := u.terraformOutputter.Get(state.TFState, "network_name")
	if err != nil {
		return err
	}
	subnetworkName, err := u.terraformOutputter.Get(state.TFState, "subnetwork_name")
	if err != nil {
		return err
	}
	boshTag, err := u.terraformOutputter.Get(state.TFState, "bosh_open_tag_name")
	if err != nil {
		return err
	}
	internalTag, err := u.terraformOutputter.Get(state.TFState, "internal_tag_name")
	if err != nil {
		return err
	}
	directorAddress, err := u.terraformOutputter.Get(state.TFState, "director_address")
	if err != nil {
		return err
	}

	_ = boshinit.InfrastructureConfiguration{
		ExternalIP: externalIP,
		GCP: boshinit.InfrastructureConfigurationGCP{
			Zone:           state.GCP.Zone,
			NetworkName:    networkName,
			SubnetworkName: subnetworkName,
			BOSHTag:        boshTag,
			InternalTag:    internalTag,
			Project:        state.GCP.ProjectID,
			JsonKey:        state.GCP.ServiceAccountKey,
		},
	}

	deployInput := bosh.DeployInput{
		IAAS:         "gcp",
		DirectorName: fmt.Sprintf("bosh-%s", state.EnvID),
		Zone:         state.GCP.Zone,
		Network:      networkName,
		Subnetwork:   subnetworkName,
		Tags: []string{
			boshTag,
			internalTag,
		},
		ProjectID:       state.GCP.ProjectID,
		ExternalIP:      externalIP,
		CredentialsJSON: state.GCP.ServiceAccountKey,
		PrivateKey:      state.KeyPair.PrivateKey,
	}
	deployOutput, err := u.boshDeployer.Deploy(deployInput)
	if err != nil {
		return err
	}

	if state.BOSH.IsEmpty() {
		directorOutputs := getDirectorOutputs(deployOutput.Variables)

		variablesYAMLContents, err := marshal(deployOutput.Variables)
		if err != nil {
			return err
		}
		variablesYAML := string(variablesYAMLContents)

		state.BOSH = storage.BOSH{
			DirectorName:           deployInput.DirectorName,
			DirectorAddress:        directorAddress,
			DirectorUsername:       DIRECTOR_USERNAME,
			DirectorPassword:       directorOutputs.directorPassword,
			DirectorSSLCA:          directorOutputs.directorSSLCA,
			DirectorSSLCertificate: directorOutputs.directorSSLCertificate,
			DirectorSSLPrivateKey:  directorOutputs.directorSSLPrivateKey,
			Variables:              variablesYAML,
		}
	}

	state.BOSH.State = deployOutput.BOSHState

	err = u.stateStore.Set(state)
	if err != nil {
		return err
	}

	boshClient := u.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword)

	cfHTTPSRouterBackendService := ""
	cfSSHProxyTargetPool := ""
	cfTCPRouterTargetPool := ""
	cfWSTargetPool := ""
	if state.LB.Type == "cf" {
		cfHTTPSRouterBackendService, err = u.terraformOutputter.Get(state.TFState, "router_backend_service")
		if err != nil {
			return err
		}

		cfSSHProxyTargetPool, err = u.terraformOutputter.Get(state.TFState, "ssh_proxy_target_pool")
		if err != nil {
			return err
		}

		cfTCPRouterTargetPool, err = u.terraformOutputter.Get(state.TFState, "tcp_router_target_pool")
		if err != nil {
			return err
		}

		cfWSTargetPool, err = u.terraformOutputter.Get(state.TFState, "ws_target_pool")
		if err != nil {
			return err
		}
	}

	concourseTargetPool := ""
	if state.LB.Type == "concourse" {
		concourseTargetPool, err = u.terraformOutputter.Get(state.TFState, "concourse_target_pool")
		if err != nil {
			return err
		}
	}

	u.logger.Step("generating cloud config")
	cloudConfig, err := u.cloudConfigGenerator.Generate(gcp.CloudConfigInput{
		AZs:                 zones,
		Tags:                []string{internalTag},
		NetworkName:         networkName,
		SubnetworkName:      subnetworkName,
		ConcourseTargetPool: concourseTargetPool,
		CFBackends: gcp.CFBackends{
			Router:    cfHTTPSRouterBackendService,
			SSHProxy:  cfSSHProxyTargetPool,
			TCPRouter: cfTCPRouterTargetPool,
			WS:        cfWSTargetPool,
		},
	})
	if err != nil {
		return err
	}

	manifestYAML, err := marshal(cloudConfig)
	if err != nil {
		return err // not tested
	}

	u.logger.Step("applying cloud config")
	if err := boshClient.UpdateCloudConfig(manifestYAML); err != nil {
		return err
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

func (u GCPUp) parseUpConfig(upConfig GCPUpConfig) (storage.GCP, error) {
	if upConfig.ServiceAccountKeyPath == "" {
		return storage.GCP{}, errors.New("GCP service account key must be provided")
	}

	sak, err := ioutil.ReadFile(upConfig.ServiceAccountKeyPath)
	if err != nil {
		return storage.GCP{}, fmt.Errorf("error reading service account key: %v", err)
	}

	var tmp interface{}
	err = json.Unmarshal(sak, &tmp)
	if err != nil {
		return storage.GCP{}, fmt.Errorf("error parsing service account key: %v", err)
	}

	return storage.GCP{
		ServiceAccountKey: string(sak),
		ProjectID:         upConfig.ProjectID,
		Zone:              upConfig.Zone,
		Region:            upConfig.Region,
	}, nil
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

func getDirectorOutputs(variables map[string]interface{}) directorOutputs {
	directorSSLInterfaceMap := variables["director_ssl"].(map[interface{}]interface{})
	directorSSL := map[string]string{}
	for k, v := range directorSSLInterfaceMap {
		directorSSL[k.(string)] = v.(string)
	}

	return directorOutputs{
		directorPassword:       variables["admin_password"].(string),
		directorSSLCA:          directorSSL["ca"],
		directorSSLCertificate: directorSSL["certificate"],
		directorSSLPrivateKey:  directorSSL["private_key"],
	}
}
