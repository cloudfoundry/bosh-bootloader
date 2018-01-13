package bosh

import (
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

var (
	osSetenv   = os.Setenv
	osUnsetenv = os.Unsetenv
)

type Manager struct {
	executor     executor
	logger       logger
	socks5Proxy  socks5Proxy
	stateStore   stateStore
	sshKeyGetter sshKeyGetter
}

type directorVars struct {
	username       string
	password       string
	sslCA          string
	sslCertificate string
	sslPrivateKey  string
}

type executor interface {
	PlanDirector(DirInput, string, string) error
	PlanJumpbox(DirInput, string, string) error
	CreateEnv(DirInput, storage.State) (string, error)
	DeleteEnv(DirInput, storage.State) error
	WriteDeploymentVars(DirInput, string) error
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

type sshKeyGetter interface {
	Get(string) (string, error)
}

func NewManager(executor executor, logger logger, socks5Proxy socks5Proxy, stateStore stateStore, sshKeyGetter sshKeyGetter) *Manager {
	return &Manager{
		executor:     executor,
		logger:       logger,
		socks5Proxy:  socks5Proxy,
		stateStore:   stateStore,
		sshKeyGetter: sshKeyGetter,
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

func (m *Manager) InitializeJumpbox(state storage.State) error {
	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()

	deploymentDir, err := m.stateStore.GetJumpboxDeploymentDir()
	if err != nil {
		return fmt.Errorf("Get deployment dir: %s", err)
	}

	iaasInputs := DirInput{
		StateDir: stateDir,
		VarsDir:  varsDir,
	}

	err = m.executor.PlanJumpbox(iaasInputs, deploymentDir, state.IAAS)
	if err != nil {
		return fmt.Errorf("Jumpbox interpolate: %s", err)
	}

	return nil
}

func (m *Manager) CreateJumpbox(state storage.State, terraformOutputs terraform.Outputs) (storage.State, error) {
	m.logger.Step("creating jumpbox")

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return storage.State{}, fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()
	osUnsetenv("BOSH_ALL_PROXY")
	dirInput := DirInput{
		Deployment: "jumpbox",
		StateDir:   stateDir,
		VarsDir:    varsDir,
	}

	err = m.executor.WriteDeploymentVars(dirInput, m.GetJumpboxDeploymentVars(state, terraformOutputs))
	if err != nil {
		return storage.State{}, fmt.Errorf("Write deployment vars: %s", err)
	}

	variables, err := m.executor.CreateEnv(dirInput, state)
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
		URL: terraformOutputs.GetString("jumpbox_url"),
	}

	m.logger.Step("starting socks5 proxy to jumpbox")
	jumpboxPrivateKey, err := m.sshKeyGetter.Get("jumpbox")
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

func (m *Manager) InitializeDirector(state storage.State) error {
	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()

	directorDeploymentDir, err := m.stateStore.GetDirectorDeploymentDir()
	if err != nil {
		return fmt.Errorf("Get deployment dir: %s", err)
	}

	iaasInputs := DirInput{
		StateDir: stateDir,
		VarsDir:  varsDir,
	}

	err = m.executor.PlanDirector(iaasInputs, directorDeploymentDir, state.IAAS)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) CreateDirector(state storage.State, terraformOutputs terraform.Outputs) (storage.State, error) {
	m.logger.Step("creating bosh director")

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return storage.State{}, fmt.Errorf("Get vars dir: %s", err)
	}

	stateDir := m.stateStore.GetStateDir()

	dirInput := DirInput{
		Deployment: "director",
		StateDir:   stateDir,
		VarsDir:    varsDir,
	}

	err = m.executor.WriteDeploymentVars(dirInput, m.GetDirectorDeploymentVars(state, terraformOutputs))
	if err != nil {
		return storage.State{}, fmt.Errorf("Write deployment vars: %s", err)
	}

	variables, err := m.executor.CreateEnv(dirInput, state)

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

	internalCIDR := terraformOutputs.GetString("internal_cidr")
	parsedInternalCIDR, err := ParseCIDRBlock(internalCIDR)
	if err != nil {
		internalCIDR = "10.0.0.0/24"
		parsedInternalCIDR, _ = ParseCIDRBlock(internalCIDR)
	}

	internalIP := terraformOutputs.GetString("director__internal_ip")
	if internalIP == "" {
		internalIP = parsedInternalCIDR.GetNthIP(6).String()
	}

	state.BOSH = storage.BOSH{
		DirectorName:           fmt.Sprintf("bosh-%s", state.EnvID),
		DirectorAddress:        fmt.Sprintf("https://%s:25555", internalIP),
		DirectorUsername:       directorVars.username,
		DirectorPassword:       directorVars.password,
		DirectorSSLCA:          directorVars.sslCA,
		DirectorSSLCertificate: directorVars.sslCertificate,
		DirectorSSLPrivateKey:  directorVars.sslPrivateKey,
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

	dirInput := DirInput{
		Deployment: "director",
		StateDir:   stateDir,
		VarsDir:    varsDir,
	}

	err = m.executor.WriteDeploymentVars(dirInput, m.GetDirectorDeploymentVars(state, terraformOutputs))
	if err != nil {
		return fmt.Errorf("Write deployment vars: %s", err)
	}

	jumpboxPrivateKey, err := m.sshKeyGetter.Get("jumpbox")
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

	err = m.executor.DeleteEnv(dirInput, state)
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

	dirInput := DirInput{
		Deployment: "jumpbox",
		StateDir:   stateDir,
		VarsDir:    varsDir,
	}

	err = m.executor.WriteDeploymentVars(dirInput, m.GetJumpboxDeploymentVars(state, terraformOutputs))
	if err != nil {
		return fmt.Errorf("Write deployment vars: %s", err)
	}

	err = m.executor.DeleteEnv(dirInput, state)
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
	allOutputs := map[string]interface{}{}
	for k, v := range terraformOutputs.Map {
		if strings.HasPrefix(k, "director__") || strings.HasPrefix(k, "jumpbox__") {
			continue
		}
		allOutputs[k] = v
	}

	for k, v := range terraformOutputs.Map {
		if strings.HasPrefix(k, "jumpbox__") {
			k = strings.Replace(k, "jumpbox__", "", 1)
			allOutputs[k] = v
		}
	}

	vars := sharedDeploymentVarsYAML{
		TerraformOutputs: allOutputs,
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

func (m *Manager) GetDirectorDeploymentVars(state storage.State, terraformOutputs terraform.Outputs) string {
	allOutputs := map[string]interface{}{}
	for k, v := range terraformOutputs.Map {
		if strings.HasPrefix(k, "director__") || strings.HasPrefix(k, "jumpbox__") {
			continue
		}
		allOutputs[k] = v
	}

	for k, v := range terraformOutputs.Map {
		if strings.HasPrefix(k, "director__") {
			k = strings.Replace(k, "director__", "", 1)
			allOutputs[k] = v
		}
	}

	vars := sharedDeploymentVarsYAML{
		TerraformOutputs: allOutputs,
	}

	return string(mustMarshal(vars))
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
		username:       "admin",
		password:       vars.AdminPassword,
		sslCA:          vars.DirectorSSL.CA,
		sslCertificate: vars.DirectorSSL.Certificate,
		sslPrivateKey:  vars.DirectorSSL.PrivateKey,
	}
}
