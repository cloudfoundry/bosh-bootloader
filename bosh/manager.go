package bosh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	yaml "gopkg.in/yaml.v2"
)

var (
	osSetenv   = os.Setenv
	osUnsetenv = os.Unsetenv
)

type managerFs interface {
	fileio.FileWriter
	fileio.TempDirer
}

type Manager struct {
	executor        executor
	logger          logger
	stateStore      stateStore
	sshKeyGetter    sshKeyGetter
	fs              managerFs
	boshCLIProvider boshCLIProvider
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
	Path() string
	Version() (string, error)
}

type logger interface {
	Step(string, ...interface{})
	Println(string)
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

func NewManager(executor executor, logger logger, stateStore stateStore, sshKeyGetter sshKeyGetter, fs deleterFs, boshCLIProvider boshCLIProvider) *Manager {
	return &Manager{
		executor:        executor,
		logger:          logger,
		stateStore:      stateStore,
		sshKeyGetter:    sshKeyGetter,
		fs:              fs,
		boshCLIProvider: boshCLIProvider,
	}
}

func (m *Manager) Path() string {
	return m.executor.Path()
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
		return err
	}

	stateDir := m.stateStore.GetStateDir()

	deploymentDir, err := m.stateStore.GetJumpboxDeploymentDir()
	if err != nil {
		return err
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
		return storage.State{}, err
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

	_, err = m.executor.CreateEnv(dirInput, state)
	if err != nil {
		return storage.State{}, NewManagerCreateError(state, err)
	}
	m.logger.Step("created jumpbox")

	state.Jumpbox = storage.Jumpbox{
		URL: terraformOutputs.GetString("jumpbox_url"),
	}

	dir, err := m.fs.TempDir("", "bosh-jumpbox")
	if err != nil {
		return storage.State{}, fmt.Errorf("Create temp dir for jumpbox private key: %s", err)
	}

	privateKeyPath := filepath.Join(dir, "bosh_jumpbox_private.key")

	privateKeyContents, err := m.sshKeyGetter.Get("jumpbox")
	if err != nil {
		return storage.State{}, fmt.Errorf("Get jumpbox private key: %s", err)
	}

	err = m.fs.WriteFile(privateKeyPath, []byte(privateKeyContents), 0600)
	if err != nil {
		return storage.State{}, fmt.Errorf("Write jumpbox private key: %s", err)
	}

	osSetenv("BOSH_ALL_PROXY", fmt.Sprintf("ssh+socks5://jumpbox@%s?private-key=%s", state.Jumpbox.URL, privateKeyPath))

	return state, nil
}

func (m *Manager) InitializeDirector(state storage.State) error {
	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return err
	}

	stateDir := m.stateStore.GetStateDir()

	directorDeploymentDir, err := m.stateStore.GetDirectorDeploymentDir()
	if err != nil {
		return err
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

func (m *Manager) CleanUpDirector(state storage.State) error {
	if state.BOSH.IsEmpty() {
		return nil
	}

	m.logger.Step("cleaning up director resources")

	boshCLI, err := m.boshCLIProvider.AuthenticatedCLI(
		state.Jumpbox,
		os.Stderr,
		state.BOSH.DirectorAddress,
		state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword,
		state.BOSH.DirectorSSLCA,
	)

	if err != nil {
		return err
	}

	return boshCLI.Run(nil, "", []string{"clean-up", "--all"})
}

func (m *Manager) CreateDirector(state storage.State, terraformOutputs terraform.Outputs) (storage.State, error) {
	m.logger.Step("creating bosh director")

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return storage.State{}, err
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
	if err != nil {
		state.BOSH = storage.BOSH{
			Variables: variables,
		}
		return storage.State{}, NewManagerCreateError(state, err)
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
	if state.BOSH.IsEmpty() {
		return nil
	}

	m.logger.Step("destroying bosh director")

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return err
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

	dir, err := m.fs.TempDir("", "bosh-jumpbox")
	if err != nil {
		return fmt.Errorf("Create temp dir for jumpbox private key: %s", err)
	}

	privateKeyPath := filepath.Join(dir, "bosh_jumpbox_private.key")

	privateKeyContents, err := m.sshKeyGetter.Get("jumpbox")
	if err != nil {
		return fmt.Errorf("Get jumpbox private key: %s", err)
	}

	err = m.fs.WriteFile(privateKeyPath, []byte(privateKeyContents), 0600)
	if err != nil {
		return fmt.Errorf("Write jumpbox private key: %s", err)
	}

	osSetenv("BOSH_ALL_PROXY", fmt.Sprintf("ssh+socks5://jumpbox@%s?private-key=%s", state.Jumpbox.URL, privateKeyPath))

	err = m.executor.DeleteEnv(dirInput, state)
	if err != nil {
		return NewManagerDeleteError(state, err)
	}

	return nil
}

func (m *Manager) DeleteJumpbox(state storage.State, terraformOutputs terraform.Outputs) error {
	if state.Jumpbox.IsEmpty() {
		return nil
	}

	m.logger.Step("destroying jumpbox")

	varsDir, err := m.stateStore.GetVarsDir()
	if err != nil {
		return err
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
	if err != nil {
		return NewManagerDeleteError(state, err)
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
