package bosh

import (
	"errors"
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/storage"
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
	iaasInputs  InterpolateInput
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
	InternalCIDR string  `yaml:"internal_cidr,omitempty"`
	InternalGW   string  `yaml:"internal_gw,omitempty"`
	InternalIP   string  `yaml:"internal_ip,omitempty"`
	DirectorName string  `yaml:"director_name,omitempty"`
	ExternalIP   string  `yaml:"external_ip,omitempty"`
	PublicIP     string  `yaml:"public_ip,omitempty"`
	AWSYAML      AWSYAML `yaml:",inline"`
	GCPYAML      GCPYAML `yaml:",inline"`
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
	PrivateKey            string   `yaml:"private_key,flow,omitempty"`
}

type GCPYAML struct {
	Zone           string   `yaml:"zone,omitempty"`
	Network        string   `yaml:"network,omitempty"`
	Subnetwork     string   `yaml:"subnetwork,omitempty"`
	Tags           []string `yaml:"tags,omitempty"`
	ProjectID      string   `yaml:"project_id,omitempty"`
	CredentialJSON string   `yaml:"gcp_credentials_json,omitempty"`
}

type executor interface {
	DirectorInterpolate(InterpolateInput) (InterpolateOutput, error)
	JumpboxInterpolate(InterpolateInput) (JumpboxInterpolateOutput, error)
	CreateEnv(CreateEnvInput) (CreateEnvOutput, error)
	DeleteEnv(DeleteEnvInput) error
	Version() (string, error)
}

type logger interface {
	Step(string, ...interface{})
	Println(string)
}

type socks5Proxy interface {
	Start(string, string) error
	Addr() string
}

func NewManager(executor executor, logger logger, socks5Proxy socks5Proxy) *Manager {
	return &Manager{
		executor:    executor,
		logger:      logger,
		socks5Proxy: socks5Proxy,
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

func (m *Manager) CreateJumpbox(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	var err error
	m.logger.Step("creating jumpbox")

	m.iaasInputs, err = generateIAASInputs(state)
	if err != nil {
		return storage.State{}, err
	}

	m.iaasInputs.JumpboxDeploymentVars, err = m.GetJumpboxDeploymentVars(state, terraformOutputs)
	if err != nil {
		return storage.State{}, err //not tested
	}
	interpolateOutputs, err := m.executor.JumpboxInterpolate(m.iaasInputs)
	if err != nil {
		return storage.State{}, err
	}

	variables, err := yaml.Marshal(interpolateOutputs.Variables)
	if err != nil {
		return storage.State{}, err
	}

	osUnsetenv("BOSH_ALL_PROXY")
	createEnvOutputs, err := m.executor.CreateEnv(CreateEnvInput{
		Manifest:  interpolateOutputs.Manifest,
		State:     state.Jumpbox.State,
		Variables: string(variables),
	})
	switch err.(type) {
	case CreateEnvError:
		ceErr := err.(CreateEnvError)
		state.Jumpbox = storage.Jumpbox{
			Enabled:   true,
			Variables: interpolateOutputs.Variables,
			State:     ceErr.BOSHState(),
			Manifest:  interpolateOutputs.Manifest,
		}
		return storage.State{}, NewManagerCreateError(state, err)
	case error:
		return storage.State{}, err
	}

	state.Jumpbox = storage.Jumpbox{
		Enabled:   true,
		Variables: interpolateOutputs.Variables,
		State:     createEnvOutputs.State,
		Manifest:  interpolateOutputs.Manifest,
		URL:       terraformOutputs["jumpbox_url"].(string),
	}

	m.logger.Step("created jumpbox")

	m.logger.Step("starting socks5 proxy to jumpbox")
	jumpboxPrivateKey, err := getJumpboxPrivateKey(interpolateOutputs.Variables)
	if err != nil {
		return storage.State{}, err
	}

	err = m.socks5Proxy.Start(jumpboxPrivateKey, state.Jumpbox.URL)
	if err != nil {
		return storage.State{}, err
	}

	osSetenv("BOSH_ALL_PROXY", fmt.Sprintf("socks5://%s", m.socks5Proxy.Addr()))

	return state, nil
}

func (m *Manager) CreateDirector(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	var err error
	var directorAddress string

	directorAddress = terraformOutputs["director_address"].(string)

	if state.Jumpbox.Enabled {
		directorAddress = fmt.Sprintf("https://%s:25555", DIRECTOR_INTERNAL_IP)
	} else {
		m.iaasInputs, err = generateIAASInputs(state)
		if err != nil {
			return storage.State{}, err
		}
	}

	m.logger.Step("creating bosh director")
	m.iaasInputs.DeploymentVars, err = m.GetDeploymentVars(state, terraformOutputs)
	if err != nil {
		return storage.State{}, err //not tested
	}

	m.iaasInputs.OpsFile = state.BOSH.UserOpsFile

	interpolateOutputs, err := m.executor.DirectorInterpolate(m.iaasInputs)
	if err != nil {
		return storage.State{}, err
	}

	createEnvOutputs, err := m.executor.CreateEnv(CreateEnvInput{
		Manifest:  interpolateOutputs.Manifest,
		State:     state.BOSH.State,
		Variables: interpolateOutputs.Variables,
	})
	switch err.(type) {
	case CreateEnvError:
		ceErr := err.(CreateEnvError)
		state.BOSH = storage.BOSH{
			Variables: interpolateOutputs.Variables,
			State:     ceErr.BOSHState(),
			Manifest:  interpolateOutputs.Manifest,
		}
		return storage.State{}, NewManagerCreateError(state, err)
	case error:
		return storage.State{}, err
	}

	directorVars, err := getDirectorVars(interpolateOutputs.Variables)
	if err != nil {
		return storage.State{}, fmt.Errorf("failed to get director outputs:\n%s", err.Error())
	}

	state.BOSH = storage.BOSH{
		DirectorName:           fmt.Sprintf("bosh-%s", state.EnvID),
		DirectorAddress:        directorAddress,
		DirectorUsername:       DIRECTOR_USERNAME,
		DirectorPassword:       directorVars.directorPassword,
		DirectorSSLCA:          directorVars.directorSSLCA,
		DirectorSSLCertificate: directorVars.directorSSLCertificate,
		DirectorSSLPrivateKey:  directorVars.directorSSLPrivateKey,
		Variables:              interpolateOutputs.Variables,
		State:                  createEnvOutputs.State,
		Manifest:               interpolateOutputs.Manifest,
	}

	m.logger.Step("created bosh director")
	return state, nil
}

func (m *Manager) Delete(state storage.State, terraformOutputs map[string]interface{}) error {
	iaasInputs, err := generateIAASInputs(state)
	if err != nil {
		return err
	}

	if state.Jumpbox.Enabled {
		jumpboxPrivateKey, err := getJumpboxPrivateKey(state.Jumpbox.Variables)
		if err != nil {
			return err
		}

		err = m.socks5Proxy.Start(jumpboxPrivateKey, state.Jumpbox.URL)
		if err != nil {
			return err
		}

		osSetenv("BOSH_ALL_PROXY", fmt.Sprintf("socks5://%s", m.socks5Proxy.Addr()))

		iaasInputs.JumpboxDeploymentVars, err = m.GetJumpboxDeploymentVars(state, terraformOutputs)
		if err != nil {
			return err //not tested
		}
	}

	iaasInputs.DeploymentVars, err = m.GetDeploymentVars(state, terraformOutputs)
	if err != nil {
		return err //not tested
	}

	iaasInputs.OpsFile = state.BOSH.UserOpsFile

	interpolateOutputs, err := m.executor.DirectorInterpolate(iaasInputs)
	if err != nil {
		return err
	}

	err = m.executor.DeleteEnv(DeleteEnvInput{
		Manifest:  interpolateOutputs.Manifest,
		State:     state.BOSH.State,
		Variables: interpolateOutputs.Variables,
	})
	switch err.(type) {
	case DeleteEnvError:
		deErr := err.(DeleteEnvError)
		state.BOSH.State = deErr.BOSHState()
		return NewManagerDeleteError(state, err)
	case error:
		return err
	}

	return nil
}

func (m *Manager) DeleteJumpbox(state storage.State, terraformOutputs map[string]interface{}) error {
	if !state.Jumpbox.Enabled {
		return nil
	}

	m.logger.Step("destroying jumpbox")
	iaasInputs, err := generateIAASInputs(state)
	if err != nil {
		return err
	}

	iaasInputs.JumpboxDeploymentVars, err = m.GetJumpboxDeploymentVars(state, terraformOutputs)
	if err != nil {
		return err //not tested
	}

	interpolateOutputs, err := m.executor.JumpboxInterpolate(iaasInputs)
	if err != nil {
		return err
	}

	err = m.executor.DeleteEnv(DeleteEnvInput{
		Manifest:  interpolateOutputs.Manifest,
		State:     state.Jumpbox.State,
		Variables: interpolateOutputs.Variables,
	})
	switch err.(type) {
	case DeleteEnvError:
		deErr := err.(DeleteEnvError)
		state.Jumpbox.State = deErr.BOSHState()
		return NewManagerDeleteError(state, err)
	case error:
		return err
	}

	return nil
}

func (m *Manager) GetJumpboxDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) (string, error) {
	gcpVars, err := yaml.Marshal(sharedDeploymentVarsYAML{
		InternalCIDR: "10.0.0.0/24",
		InternalGW:   "10.0.0.1",
		InternalIP:   "10.0.0.5",
		DirectorName: fmt.Sprintf("bosh-%s", state.EnvID),
		ExternalIP:   getTerraformOutput("external_ip", terraformOutputs),
		GCPYAML: GCPYAML{
			Zone:           state.GCP.Zone,
			Network:        getTerraformOutput("network_name", terraformOutputs),
			Subnetwork:     getTerraformOutput("subnetwork_name", terraformOutputs),
			Tags:           []string{getTerraformOutput("bosh_open_tag_name", terraformOutputs)},
			ProjectID:      state.GCP.ProjectID,
			CredentialJSON: state.GCP.ServiceAccountKey,
		},
	})
	if err != nil {
		panic(err)
	}

	return string(gcpVars), nil
}

func getTerraformOutput(key string, outputs map[string]interface{}) string {
	if value, ok := outputs[key]; ok {
		return fmt.Sprintf("%s", value)
	}
	return ""
}

func (m *Manager) GetDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) (string, error) {
	var vars []byte

	switch state.IAAS {
	case "gcp":
		if state.Jumpbox.Enabled {
			vars, _ = yaml.Marshal(sharedDeploymentVarsYAML{
				InternalCIDR: "10.0.0.0/24",
				InternalGW:   "10.0.0.1",
				InternalIP:   DIRECTOR_INTERNAL_IP,
				DirectorName: fmt.Sprintf("bosh-%s", state.EnvID),
				GCPYAML: GCPYAML{
					Zone:           state.GCP.Zone,
					Network:        getTerraformOutput("network_name", terraformOutputs),
					Subnetwork:     getTerraformOutput("subnetwork_name", terraformOutputs),
					Tags:           []string{getTerraformOutput("bosh_director_tag_name", terraformOutputs)},
					ProjectID:      state.GCP.ProjectID,
					CredentialJSON: state.GCP.ServiceAccountKey,
				},
			})
		} else {
			vars, _ = yaml.Marshal(sharedDeploymentVarsYAML{
				InternalCIDR: "10.0.0.0/24",
				InternalGW:   "10.0.0.1",
				InternalIP:   DIRECTOR_INTERNAL_IP,
				DirectorName: fmt.Sprintf("bosh-%s", state.EnvID),
				ExternalIP:   getTerraformOutput("external_ip", terraformOutputs),
				PublicIP:     getTerraformOutput("external_ip", terraformOutputs),
				GCPYAML: GCPYAML{
					Zone:           state.GCP.Zone,
					Network:        getTerraformOutput("network_name", terraformOutputs),
					Subnetwork:     getTerraformOutput("subnetwork_name", terraformOutputs),
					Tags:           []string{getTerraformOutput("bosh_director_tag_name", terraformOutputs), getTerraformOutput("bosh_open_tag_name", terraformOutputs)},
					ProjectID:      state.GCP.ProjectID,
					CredentialJSON: state.GCP.ServiceAccountKey,
				},
			})
		}
	case "aws":
		vars, _ = yaml.Marshal(sharedDeploymentVarsYAML{
			InternalCIDR: "10.0.0.0/24",
			InternalGW:   "10.0.0.1",
			InternalIP:   DIRECTOR_INTERNAL_IP,
			ExternalIP:   getTerraformOutput("external_ip", terraformOutputs),
			PublicIP:     getTerraformOutput("external_ip", terraformOutputs),
			DirectorName: fmt.Sprintf("bosh-%s", state.EnvID),
			AWSYAML: AWSYAML{
				AZ:                    getTerraformOutput("bosh_subnet_availability_zone", terraformOutputs),
				SubnetID:              getTerraformOutput("bosh_subnet_id", terraformOutputs),
				AccessKeyID:           state.AWS.AccessKeyID,
				SecretAccessKey:       state.AWS.SecretAccessKey,
				IAMInstanceProfile:    getTerraformOutput("bosh_iam_instance_profile", terraformOutputs),
				DefaultKeyName:        getTerraformOutput("bosh_vms_key_name", terraformOutputs),
				DefaultSecurityGroups: []string{getTerraformOutput("bosh_security_group", terraformOutputs)},
				Region:                state.AWS.Region,
				PrivateKey:            getTerraformOutput("bosh_vms_private_key", terraformOutputs),
			},
		})
	}

	return string(vars), nil
}

func generateIAASInputs(state storage.State) (InterpolateInput, error) {
	switch state.IAAS {
	case "gcp", "aws":
		return InterpolateInput{
			IAAS:      state.IAAS,
			BOSHState: state.BOSH.State,
			Variables: state.BOSH.Variables,
		}, nil
	default:
		return InterpolateInput{}, errors.New("A valid IAAS was not provided")
	}
}

func getJumpboxPrivateKey(v string) (string, error) {
	variables := map[string]interface{}{}

	err := yaml.Unmarshal([]byte(v), &variables)
	if err != nil {
		return "", err
	}

	jumpboxMap := variables["jumpbox_ssh"].(map[interface{}]interface{})
	jumpboxSSH := map[string]string{}
	for k, v := range jumpboxMap {
		jumpboxSSH[k.(string)] = v.(string)
	}

	return jumpboxSSH["private_key"], nil
}

func getDirectorVars(v string) (directorVars, error) {
	variables := map[string]interface{}{}

	err := yaml.Unmarshal([]byte(v), &variables)
	if err != nil {
		return directorVars{}, err
	}

	directorSSLInterfaceMap := variables["director_ssl"].(map[interface{}]interface{})
	directorSSL := map[string]string{}
	for k, v := range directorSSLInterfaceMap {
		directorSSL[k.(string)] = v.(string)
	}

	return directorVars{
		directorPassword:       variables["admin_password"].(string),
		directorSSLCA:          directorSSL["ca"],
		directorSSLCertificate: directorSSL["certificate"],
		directorSSLPrivateKey:  directorSSL["private_key"],
	}, nil
}
