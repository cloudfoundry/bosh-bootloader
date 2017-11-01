package bosh

import (
	"errors"
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

var (
	osSetenv   = os.Setenv
	osUnsetenv = os.Unsetenv
)

const (
	DIRECTOR_USERNAME    = "admin"
	DIRECTOR_INTERNAL_IP = "10.0.0.6"
)

type Manager struct {
	executor    executor
	logger      logger
	socks5Proxy socks5Proxy
	stateStore  stateStore
}

type directorVars struct {
	directorPassword       string
	directorSSLCA          string
	directorSSLCertificate string
	directorSSLPrivateKey  string
}

type deploymentVariables struct {
	DirectorName          string
	Zone                  string
	Network               string
	Subnetwork            string
	Tags                  []string
	ProjectID             string
	ExternalIP            string
	CredentialsJSON       string
	PrivateKey            string
	DefaultKeyName        string
	DefaultSecurityGroups []string
	SubnetID              string
	AZ                    string
	Region                string
	SecretAccessKey       string
	AccessKeyID           string
}

type sharedDeploymentVarsYAML struct {
	InternalCIDR string    `yaml:"internal_cidr,omitempty"`
	InternalGW   string    `yaml:"internal_gw,omitempty"`
	InternalIP   string    `yaml:"internal_ip,omitempty"`
	DirectorName string    `yaml:"director_name,omitempty"`
	ExternalIP   string    `yaml:"external_ip,omitempty"`
	PrivateKey   string    `yaml:"private_key,flow,omitempty"`
	AWSYAML      AWSYAML   `yaml:",inline"`
	GCPYAML      GCPYAML   `yaml:",inline"`
	AzureYAML    AzureYAML `yaml:",inline"`
}

type AWSYAML struct {
	AZ                    string   `yaml:"az,omitempty"`
	SubnetID              string   `yaml:"subnet_id,omitempty"`
	AccessKeyID           string   `yaml:"access_key_id,omitempty"`
	SecretAccessKey       string   `yaml:"secret_access_key,omitempty"`
	IAMInstanceProfile    string   `yaml:"iam_instance_profile,omitempty"`
	DefaultKeyName        string   `yaml:"default_key_name,omitempty"`
	DefaultSecurityGroups []string `yaml:"default_security_groups,omitempty"`
	Region                string   `yaml:"region,omitempty"`
	KMSKeyARN             string   `yaml:"kms_key_arn,omitempty"`
}

type GCPYAML struct {
	Zone           string   `yaml:"zone,omitempty"`
	Network        string   `yaml:"network,omitempty"`
	Subnetwork     string   `yaml:"subnetwork,omitempty"`
	Tags           []string `yaml:"tags,omitempty"`
	ProjectID      string   `yaml:"project_id,omitempty"`
	CredentialJSON string   `yaml:"gcp_credentials_json,omitempty"`
}

type AzureYAML struct {
	VNetName             string `yaml:"vnet_name,omitempty"`
	SubnetName           string `yaml:"subnet_name,omitempty"`
	SubscriptionID       string `yaml:"subscription_id,omitempty"`
	TenantID             string `yaml:"tenant_id,omitempty"`
	ClientID             string `yaml:"client_id,omitempty"`
	ClientSecret         string `yaml:"client_secret,omitempty"`
	ResourceGroupName    string `yaml:"resource_group_name,omitempty"`
	StorageAccountName   string `yaml:"storage_account_name,omitempty"`
	DefaultSecurityGroup string `yaml:"default_security_group,omitempty"`
	PublicKey            string `yaml:"public_key,flow,omitempty"`
}

type executor interface {
	DirectorCreateEnvArgs(InterpolateInput) error
	JumpboxCreateEnvArgs(InterpolateInput) error
	CreateEnv(CreateEnvInput) (string, error)
	DeleteEnv(DeleteEnvInput) error
	Version() (string, error)
}

type logger interface {
	Step(string, ...interface{})
	Println(string)
}

type socks5Proxy interface {
	Start(string, string) error
	Addr() (string, error)
}

type stateStore interface {
	GetStateDir() string
	GetVarsDir() (string, error)
	GetDirectorDeploymentDir() (string, error)
	GetJumpboxDeploymentDir() (string, error)
}

func NewManager(executor executor, logger logger, socks5Proxy socks5Proxy, stateStore stateStore) *Manager {
	return &Manager{
		executor:    executor,
		logger:      logger,
		socks5Proxy: socks5Proxy,
		stateStore:  stateStore,
	}
}

func (m *Manager) Version() (string, error) {
	version, err := m.executor.Version()
	switch err.(type) {
	case BOSHVersionError:
		m.logger.Println("warning: BOSH version could not be parsed")
	}
	return version, err
}

func (m *Manager) InitializeJumpbox(state storage.State, terraformOutputs terraform.Outputs) error {
	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()

	deploymentDir, err := m.stateStore.GetJumpboxDeploymentDir()
	if err != nil {
		return fmt.Errorf("Get deployment dir: %s", err)
	}

	iaasInputs := InterpolateInput{
		DeploymentDir:  deploymentDir,
		StateDir:       stateDir,
		VarsDir:        varsDir,
		IAAS:           state.IAAS,
		DeploymentVars: m.GetJumpboxDeploymentVars(state, terraformOutputs),
		Variables:      state.Jumpbox.Variables,
		BOSHState:      state.Jumpbox.State,
	}

	err = m.executor.JumpboxCreateEnvArgs(iaasInputs)
	if err != nil {
		return fmt.Errorf("Jumpbox interpolate: %s", err)
	}

	return nil
}

func (m *Manager) CreateJumpbox(state storage.State, jumpboxURL string) (storage.State, error) {
	m.logger.Step("creating jumpbox")

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return storage.State{}, fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()
	osUnsetenv("BOSH_ALL_PROXY")
	variables, err := m.executor.CreateEnv(CreateEnvInput{
		Deployment: "jumpbox",
		VarsDir:    varsDir,
		StateDir:   stateDir,
	})
	switch err.(type) {
	case CreateEnvError:
		ceErr := err.(CreateEnvError)
		state.Jumpbox = storage.Jumpbox{
			Variables: variables,
			State:     ceErr.BOSHState(),
		}
		return storage.State{}, fmt.Errorf("Create jumpbox env: %s", NewManagerCreateError(state, err))
	case error:
		return storage.State{}, fmt.Errorf("Create jumpbox env: %s", err)
	}
	m.logger.Step("created jumpbox")

	state.Jumpbox = storage.Jumpbox{
		Variables: variables,
		URL:       jumpboxURL,
	}

	m.logger.Step("starting socks5 proxy to jumpbox")
	jumpboxPrivateKey, err := getJumpboxPrivateKey(variables)
	if err != nil {
		return storage.State{}, fmt.Errorf("jumpbox key: %s", err)
	}

	err = m.socks5Proxy.Start(jumpboxPrivateKey, state.Jumpbox.URL)
	if err != nil {
		return storage.State{}, fmt.Errorf("Start proxy: %s", err)
	}

	addr, err := m.socks5Proxy.Addr()
	if err != nil {
		return storage.State{}, fmt.Errorf("Get proxy address: %s", err)
	}
	osSetenv("BOSH_ALL_PROXY", fmt.Sprintf("socks5://%s", addr))

	m.logger.Step("started proxy")
	return state, nil
}

func (m *Manager) InitializeDirector(state storage.State, terraformOutputs terraform.Outputs) error {
	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()

	directorDeploymentDir, err := m.stateStore.GetDirectorDeploymentDir()
	if err != nil {
		return fmt.Errorf("Get deployment dir: %s", err)
	}

	iaasInputs := InterpolateInput{
		DeploymentDir:  directorDeploymentDir,
		StateDir:       stateDir,
		VarsDir:        varsDir,
		IAAS:           state.IAAS,
		DeploymentVars: m.GetDirectorDeploymentVars(state, terraformOutputs),
		Variables:      state.BOSH.Variables,
		OpsFile:        state.BOSH.UserOpsFile,
		BOSHState:      state.BOSH.State,
	}

	err = m.executor.DirectorCreateEnvArgs(iaasInputs)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) CreateDirector(state storage.State) (storage.State, error) {
	m.logger.Step("creating bosh director")

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return storage.State{}, fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()

	variables, err := m.executor.CreateEnv(CreateEnvInput{
		Deployment: "director",
		StateDir:   stateDir,
		VarsDir:    varsDir,
	})

	switch err.(type) {
	case CreateEnvError:
		ceErr := err.(CreateEnvError)
		state.BOSH = storage.BOSH{
			Variables: variables,
			State:     ceErr.BOSHState(),
		}
		return storage.State{}, NewManagerCreateError(state, err)
	case error:
		return storage.State{}, fmt.Errorf("Create director env: %s", err)
	}

	directorVars := getDirectorVars(variables)

	state.BOSH = storage.BOSH{
		DirectorName:           fmt.Sprintf("bosh-%s", state.EnvID),
		DirectorAddress:        fmt.Sprintf("https://%s:25555", DIRECTOR_INTERNAL_IP),
		DirectorUsername:       DIRECTOR_USERNAME,
		DirectorPassword:       directorVars.directorPassword,
		DirectorSSLCA:          directorVars.directorSSLCA,
		DirectorSSLCertificate: directorVars.directorSSLCertificate,
		DirectorSSLPrivateKey:  directorVars.directorSSLPrivateKey,
		Variables:              variables,
		UserOpsFile:            state.BOSH.UserOpsFile,
	}

	m.logger.Step("created bosh director")
	return state, nil
}

func (m *Manager) DeleteDirector(state storage.State, terraformOutputs terraform.Outputs) error {
	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()

	deploymentDir, err := m.stateStore.GetDirectorDeploymentDir()
	if err != nil {
		return fmt.Errorf("Get deployment dir: %s", err)
	}

	iaasInputs := InterpolateInput{
		DeploymentDir: deploymentDir,
		StateDir:      stateDir,
		VarsDir:       varsDir,
		IAAS:          state.IAAS,
		BOSHState:     state.BOSH.State,
		Variables:     state.BOSH.Variables,
		OpsFile:       state.BOSH.UserOpsFile,
	}

	jumpboxPrivateKey, err := getJumpboxPrivateKey(state.Jumpbox.Variables)
	if err != nil {
		return fmt.Errorf("Delete bosh director: %s", err)
	}

	err = m.socks5Proxy.Start(jumpboxPrivateKey, state.Jumpbox.URL)
	if err != nil {
		return fmt.Errorf("Start socks5 proxy: %s", err)
	}

	addr, err := m.socks5Proxy.Addr()
	if err != nil {
		return fmt.Errorf("Get proxy address: %s", err)
	}
	osSetenv("BOSH_ALL_PROXY", fmt.Sprintf("socks5://%s", addr))

	iaasInputs.DeploymentVars = m.GetDirectorDeploymentVars(state, terraformOutputs)

	err = m.executor.DirectorCreateEnvArgs(iaasInputs)
	if err != nil {
		return err
	}

	err = m.executor.DeleteEnv(DeleteEnvInput{
		Deployment: "director",
		VarsDir:    varsDir,
		StateDir:   stateDir,
	})
	switch err.(type) {
	case DeleteEnvError:
		deErr := err.(DeleteEnvError)
		state.BOSH.State = deErr.BOSHState()
		return NewManagerDeleteError(state, err)
	case error:
		return fmt.Errorf("Delete director env: %s", err)
	}

	return nil
}

func (m *Manager) DeleteJumpbox(state storage.State, terraformOutputs terraform.Outputs) error {
	m.logger.Step("destroying jumpbox")

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()

	deploymentDir, err := m.stateStore.GetJumpboxDeploymentDir()
	if err != nil {
		return fmt.Errorf("Get deployment dir: %s", err)
	}

	iaasInputs := InterpolateInput{
		DeploymentDir:  deploymentDir,
		StateDir:       stateDir,
		VarsDir:        varsDir,
		IAAS:           state.IAAS,
		Variables:      state.Jumpbox.Variables,
		DeploymentVars: m.GetJumpboxDeploymentVars(state, terraformOutputs),
	}

	err = m.executor.JumpboxCreateEnvArgs(iaasInputs)
	if err != nil {
		return err
	}

	err = m.executor.DeleteEnv(DeleteEnvInput{
		Deployment: "jumpbox",
		StateDir:   stateDir,
		VarsDir:    varsDir,
	})
	switch err.(type) {
	case DeleteEnvError:
		deErr := err.(DeleteEnvError)
		state.Jumpbox.State = deErr.BOSHState()
		return NewManagerDeleteError(state, err)
	case error:
		return fmt.Errorf("Delete jumpbox env: %s", err)
	}

	return nil
}

func (m *Manager) GetJumpboxDeploymentVars(state storage.State, terraformOutputs terraform.Outputs) string {
	vars := sharedDeploymentVarsYAML{
		InternalCIDR: "10.0.0.0/24",
		InternalGW:   "10.0.0.1",
		InternalIP:   "10.0.0.5",
		DirectorName: fmt.Sprintf("bosh-%s", state.EnvID),
		ExternalIP:   terraformOutputs.GetString("external_ip"),
	}

	switch state.IAAS {
	case "gcp":
		vars.GCPYAML = GCPYAML{
			Zone:           state.GCP.Zone,
			Network:        terraformOutputs.GetString("network_name"),
			Subnetwork:     terraformOutputs.GetString("subnetwork_name"),
			Tags:           []string{terraformOutputs.GetString("bosh_open_tag_name"), terraformOutputs.GetString("jumpbox_tag_name")},
			ProjectID:      state.GCP.ProjectID,
			CredentialJSON: state.GCP.ServiceAccountKey,
		}
	case "aws":
		vars.AWSYAML = AWSYAML{
			AZ:                    terraformOutputs.GetString("bosh_subnet_availability_zone"),
			SubnetID:              terraformOutputs.GetString("bosh_subnet_id"),
			AccessKeyID:           state.AWS.AccessKeyID,
			SecretAccessKey:       state.AWS.SecretAccessKey,
			IAMInstanceProfile:    terraformOutputs.GetString("bosh_iam_instance_profile"),
			DefaultKeyName:        terraformOutputs.GetString("bosh_vms_key_name"),
			DefaultSecurityGroups: []string{terraformOutputs.GetString("jumpbox_security_group")},
			Region:                state.AWS.Region,
		}
		vars.PrivateKey = terraformOutputs.GetString("bosh_vms_private_key")
	case "azure":
		vars.AzureYAML = AzureYAML{
			VNetName:             terraformOutputs.GetString("bosh_network_name"),
			SubnetName:           terraformOutputs.GetString("bosh_subnet_name"),
			SubscriptionID:       state.Azure.SubscriptionID,
			TenantID:             state.Azure.TenantID,
			ClientID:             state.Azure.ClientID,
			ClientSecret:         state.Azure.ClientSecret,
			ResourceGroupName:    terraformOutputs.GetString("bosh_resource_group_name"),
			StorageAccountName:   terraformOutputs.GetString("bosh_storage_account_name"),
			DefaultSecurityGroup: terraformOutputs.GetString("bosh_default_security_group"),
			PublicKey:            terraformOutputs.GetString("bosh_vms_public_key"),
		}
		vars.PrivateKey = terraformOutputs.GetString("bosh_vms_private_key")
	}

	return string(mustMarshal(vars))
}

func mustMarshal(yamlStruct interface{}) []byte {
	yamlBytes, err := yaml.Marshal(yamlStruct)
	if err != nil {
		// this should never happen since we are constructing the YAML to be marshaled
		panic("bosh manager: marshal yaml: unexpected error")
	}
	return yamlBytes
}

func getTerraformOutput(key string, outputs terraform.Outputs) string {
	if value, ok := outputs.Map[key]; ok {
		return fmt.Sprintf("%s", value)
	}
	return ""
}

func (m *Manager) GetDirectorDeploymentVars(state storage.State, terraformOutputs terraform.Outputs) string {
	vars := sharedDeploymentVarsYAML{
		InternalCIDR: "10.0.0.0/24",
		InternalGW:   "10.0.0.1",
		InternalIP:   DIRECTOR_INTERNAL_IP,
		DirectorName: fmt.Sprintf("bosh-%s", state.EnvID),
	}

	switch state.IAAS {
	case "gcp":
		vars.GCPYAML = GCPYAML{
			Zone:           state.GCP.Zone,
			Network:        terraformOutputs.GetString("network_name"),
			Subnetwork:     terraformOutputs.GetString("subnetwork_name"),
			Tags:           []string{terraformOutputs.GetString("bosh_director_tag_name")},
			ProjectID:      state.GCP.ProjectID,
			CredentialJSON: state.GCP.ServiceAccountKey,
		}
	case "aws":
		vars.AWSYAML = AWSYAML{
			AZ:                    terraformOutputs.GetString("bosh_subnet_availability_zone"),
			SubnetID:              terraformOutputs.GetString("bosh_subnet_id"),
			AccessKeyID:           state.AWS.AccessKeyID,
			SecretAccessKey:       state.AWS.SecretAccessKey,
			IAMInstanceProfile:    terraformOutputs.GetString("bosh_iam_instance_profile"),
			DefaultKeyName:        terraformOutputs.GetString("bosh_vms_key_name"),
			DefaultSecurityGroups: []string{terraformOutputs.GetString("bosh_security_group")},
			Region:                state.AWS.Region,
			KMSKeyARN:             terraformOutputs.GetString("kms_key_arn"),
		}
		vars.PrivateKey = terraformOutputs.GetString("bosh_vms_private_key")
	case "azure":
		vars.AzureYAML = AzureYAML{
			VNetName:             terraformOutputs.GetString("bosh_network_name"),
			SubnetName:           terraformOutputs.GetString("bosh_subnet_name"),
			SubscriptionID:       state.Azure.SubscriptionID,
			TenantID:             state.Azure.TenantID,
			ClientID:             state.Azure.ClientID,
			ClientSecret:         state.Azure.ClientSecret,
			ResourceGroupName:    terraformOutputs.GetString("bosh_resource_group_name"),
			StorageAccountName:   terraformOutputs.GetString("bosh_storage_account_name"),
			DefaultSecurityGroup: terraformOutputs.GetString("bosh_default_security_group"),
		}
	}

	return string(mustMarshal(vars))
}

func getJumpboxPrivateKey(v string) (string, error) {
	var vars struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	err := yaml.Unmarshal([]byte(v), &vars)
	if err != nil {
		return "", err
	}
	if vars.JumpboxSSH.PrivateKey == "" {
		return "", errors.New("cannot start proxy due to missing jumpbox private key")
	}

	return vars.JumpboxSSH.PrivateKey, nil
}

func getDirectorVars(v string) directorVars {
	var vars struct {
		AdminPassword string `yaml:"admin_password"`
		DirectorSSL   struct {
			CA          string `yaml:"ca"`
			Certificate string `yaml:"certificate"`
			PrivateKey  string `yaml:"private_key"`
		} `yaml:"director_ssl"`
	}

	err := yaml.Unmarshal([]byte(v), &vars)
	if err != nil {
		panic(err) // can't happen
	}

	return directorVars{
		directorPassword:       vars.AdminPassword,
		directorSSLCA:          vars.DirectorSSL.CA,
		directorSSLCertificate: vars.DirectorSSL.Certificate,
		directorSSLPrivateKey:  vars.DirectorSSL.PrivateKey,
	}
}
