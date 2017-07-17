package bosh

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

func NewManager(executor executor, logger logger, socks5Proxy socks5Proxy) Manager {
	return Manager{
		executor:    executor,
		logger:      logger,
		socks5Proxy: socks5Proxy,
	}
}

func (m Manager) Version() (string, error) {
	version, err := m.executor.Version()
	switch err.(type) {
	case BOSHVersionError:
		m.logger.Println("warning: BOSH version could not be parsed")
	}
	return version, err
}

func (m Manager) Create(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	var err error

	if state.Jumpbox.Enabled {
		state, err = m.createJumpbox(state, terraformOutputs)
		if err != nil {
			return storage.State{}, err
		}
	}

	state, err = m.createDirector(state, terraformOutputs)
	if err != nil {
		return storage.State{}, err
	}

	return state, nil
}

func (m Manager) Delete(state storage.State, terraformOutputs map[string]interface{}) error {
	iaasInputs, err := m.generateIAASInputs(state)
	if err != nil {
		return err
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

func (m Manager) GetJumpboxDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) (string, error) {
	vars := strings.Join([]string{
		"internal_cidr: 10.0.0.0/24",
		"internal_gw: 10.0.0.1",
		"internal_ip: 10.0.0.5",
		fmt.Sprintf("director_name: %s", fmt.Sprintf("bosh-%s", state.EnvID)),
		fmt.Sprintf("external_ip: %s", terraformOutputs["external_ip"]),
		fmt.Sprintf("zone: %s", state.GCP.Zone),
		fmt.Sprintf("network: %s", terraformOutputs["network_name"]),
		fmt.Sprintf("subnetwork: %s", terraformOutputs["subnetwork_name"]),
		fmt.Sprintf("tags: [%s, %s]", terraformOutputs["bosh_open_tag_name"], terraformOutputs["internal_tag_name"]),
		fmt.Sprintf("project_id: %s", state.GCP.ProjectID),
		fmt.Sprintf("gcp_credentials_json: '%s'", state.GCP.ServiceAccountKey),
	}, "\n")

	return strings.TrimSuffix(vars, "\n"), nil
}

func (m Manager) GetDeploymentVars(state storage.State, terraformOutputs map[string]interface{}) (string, error) {
	var vars string

	switch state.IAAS {
	case "gcp":
		if state.Jumpbox.Enabled {
			vars = strings.Join([]string{
				"internal_cidr: 10.0.0.0/24",
				"internal_gw: 10.0.0.1",
				fmt.Sprintf("internal_ip: %s", DIRECTOR_INTERNAL_IP),
				fmt.Sprintf("director_name: %s", fmt.Sprintf("bosh-%s", state.EnvID)),
				fmt.Sprintf("zone: %s", state.GCP.Zone),
				fmt.Sprintf("network: %s", terraformOutputs["network_name"]),
				fmt.Sprintf("subnetwork: %s", terraformOutputs["subnetwork_name"]),
				fmt.Sprintf("tags: [%s]", terraformOutputs["internal_tag_name"]),
				fmt.Sprintf("project_id: %s", state.GCP.ProjectID),
				fmt.Sprintf("gcp_credentials_json: '%s'", state.GCP.ServiceAccountKey),
			}, "\n")
		} else {
			vars = strings.Join([]string{
				"internal_cidr: 10.0.0.0/24",
				"internal_gw: 10.0.0.1",
				fmt.Sprintf("internal_ip: %s", DIRECTOR_INTERNAL_IP),
				fmt.Sprintf("director_name: %s", fmt.Sprintf("bosh-%s", state.EnvID)),
				fmt.Sprintf("external_ip: %s", terraformOutputs["external_ip"]),
				fmt.Sprintf("zone: %s", state.GCP.Zone),
				fmt.Sprintf("network: %s", terraformOutputs["network_name"]),
				fmt.Sprintf("subnetwork: %s", terraformOutputs["subnetwork_name"]),
				fmt.Sprintf("tags: [%s, %s]", terraformOutputs["bosh_open_tag_name"], terraformOutputs["internal_tag_name"]),
				fmt.Sprintf("project_id: %s", state.GCP.ProjectID),
				fmt.Sprintf("gcp_credentials_json: '%s'", state.GCP.ServiceAccountKey),
			}, "\n")
		}
	case "aws":
		vars = strings.Join([]string{
			"internal_cidr: 10.0.0.0/24",
			"internal_gw: 10.0.0.1",
			fmt.Sprintf("internal_ip: %s", DIRECTOR_INTERNAL_IP),
			fmt.Sprintf("director_name: %s", fmt.Sprintf("bosh-%s", state.EnvID)),
			fmt.Sprintf("external_ip: %s", terraformOutputs["external_ip"]),
			fmt.Sprintf("az: %s", terraformOutputs["az"]),
			fmt.Sprintf("subnet_id: %s", terraformOutputs["subnet_id"]),
			fmt.Sprintf("access_key_id: %s", terraformOutputs["access_key_id"]),
			fmt.Sprintf("secret_access_key: %s", terraformOutputs["secret_access_key"]),
			fmt.Sprintf("default_key_name: %s", state.KeyPair.Name),
			fmt.Sprintf("default_security_groups: [%s]", terraformOutputs["default_security_groups"]),
			fmt.Sprintf("region: %s", state.AWS.Region),
			fmt.Sprintf("private_key: |-\n  %s", strings.Replace(state.KeyPair.PrivateKey, "\n", "\n  ", -1)),
		}, "\n")
	}

	return strings.TrimSuffix(vars, "\n"), nil
}

func (m Manager) generateIAASInputs(state storage.State) (InterpolateInput, error) {
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

func (m *Manager) createJumpbox(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	var err error
	m.logger.Step("creating jumpbox")

	m.iaasInputs, err = m.generateIAASInputs(state)
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

func (m Manager) createDirector(state storage.State, terraformOutputs map[string]interface{}) (storage.State, error) {
	var err error
	var directorAddress string

	directorAddress = terraformOutputs["director_address"].(string)

	if state.Jumpbox.Enabled {
		directorAddress = fmt.Sprintf("https://%s:25555", DIRECTOR_INTERNAL_IP)
	} else {
		m.iaasInputs, err = m.generateIAASInputs(state)
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
