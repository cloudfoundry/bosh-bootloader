package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	flags "github.com/jessevdk/go-flags"
)

type logger interface {
	Println(string)
}

type StateBootstrap interface {
	GetState(string) (storage.State, error)
}

type migrator interface {
	Migrate(storage.State) (storage.State, error)
}

type fs interface {
	fileio.Stater
	fileio.TempFiler
	fileio.FileReader
	fileio.FileWriter
}

func NewConfig(bootstrap StateBootstrap, migrator migrator, logger logger, fs fs) Config {
	return Config{
		stateBootstrap: bootstrap,
		migrator:       migrator,
		logger:         logger,
		fs:             fs,
	}
}

type Config struct {
	stateBootstrap StateBootstrap
	migrator       migrator
	logger         logger
	fs             fs
}

func ParseArgs(args []string) (globalFlags, []string, error) {
	var globals globalFlags
	parser := flags.NewParser(&globals, flags.IgnoreUnknown)

	remainingArgs, err := parser.ParseArgs(args[1:])
	if err != nil {
		return globalFlags{}, remainingArgs, err
	}

	if !filepath.IsAbs(globals.StateDir) {
		workingDir, err := os.Getwd()
		if err != nil {
			return globalFlags{}, remainingArgs, err // not tested
		}
		globals.StateDir = filepath.Join(workingDir, globals.StateDir)
	}

	return globals, remainingArgs, nil
}

func (c Config) Bootstrap(args []string) (application.Configuration, error) {
	if len(args) == 1 {
		return application.Configuration{
			Command: "help",
		}, nil
	}

	globalFlags, remainingArgs, err := ParseArgs(args)
	if err != nil {
		return application.Configuration{}, err
	}

	var command string
	if len(remainingArgs) > 0 {
		command = remainingArgs[0]
	}

	if globalFlags.Version || command == "version" {
		command = "version"
		return application.Configuration{
			ShowCommandHelp: globalFlags.Help,
			Command:         command,
		}, nil
	}

	if len(remainingArgs) == 0 {
		return application.Configuration{
			Command: "help",
		}, nil
	}

	if len(remainingArgs) == 1 && command == "help" {
		return application.Configuration{
			Command: command,
		}, nil
	}

	if command == "help" {
		return application.Configuration{
			ShowCommandHelp: true,
			Command:         remainingArgs[1],
		}, nil
	}

	if globalFlags.Help {
		return application.Configuration{
			ShowCommandHelp: true,
			Command:         command,
		}, nil
	}

	state, err := c.stateBootstrap.GetState(globalFlags.StateDir)
	if err != nil {
		return application.Configuration{}, err
	}

	state, err = c.migrator.Migrate(state)
	if err != nil {
		return application.Configuration{}, err
	}

	state, err = c.updateIAASState(globalFlags, state)
	if err != nil {
		return application.Configuration{}, err
	}

	return application.Configuration{
		Global: application.GlobalConfiguration{
			Debug:    globalFlags.Debug,
			StateDir: globalFlags.StateDir,
		},
		State:           state,
		Command:         command,
		SubcommandFlags: remainingArgs[1:],
		ShowCommandHelp: globalFlags.Help,
	}, nil
}

func (c Config) updateIAASState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	if globalFlags.IAAS != "" {
		if state.IAAS != "" && globalFlags.IAAS != state.IAAS {
			return storage.State{}, fmt.Errorf("The iaas type cannot be changed for an existing environment. The current iaas type is %s.", state.IAAS)
		}
		state.IAAS = globalFlags.IAAS
	}

	switch state.IAAS {
	case "aws":
		return c.updateAWSState(globalFlags, state)
	case "azure":
		return c.updateAzureState(globalFlags, state)
	case "gcp":
		return c.updateGCPState(globalFlags, state)
	case "vsphere":
		return c.updateVSphereState(globalFlags, state)
	case "openstack":
		return c.updateOpenStackState(globalFlags, state)
	}

	return state, nil
}

func copyFlagToState(source string, sink *string) {
	if source != "" {
		*sink = source
	}
}

func (c Config) updateOpenStackState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	copyFlagToState(globalFlags.OpenStackInternalCidr, &state.OpenStack.InternalCidr)
	copyFlagToState(globalFlags.OpenStackExternalIP, &state.OpenStack.ExternalIP)
	copyFlagToState(globalFlags.OpenStackAuthURL, &state.OpenStack.AuthURL)
	copyFlagToState(globalFlags.OpenStackAZ, &state.OpenStack.AZ)
	copyFlagToState(globalFlags.OpenStackDefaultKeyName, &state.OpenStack.DefaultKeyName)
	copyFlagToState(globalFlags.OpenStackDefaultSecurityGroup, &state.OpenStack.DefaultSecurityGroup)
	copyFlagToState(globalFlags.OpenStackNetworkID, &state.OpenStack.NetworkID)
	copyFlagToState(globalFlags.OpenStackPassword, &state.OpenStack.Password)
	copyFlagToState(globalFlags.OpenStackUsername, &state.OpenStack.Username)
	copyFlagToState(globalFlags.OpenStackProject, &state.OpenStack.Project)
	copyFlagToState(globalFlags.OpenStackDomain, &state.OpenStack.Domain)
	copyFlagToState(globalFlags.OpenStackRegion, &state.OpenStack.Region)
	copyFlagToState(globalFlags.OpenStackRegion, &state.OpenStack.Region)

	if globalFlags.OpenStackPrivateKey != "" {
		absKeyPath, err := filepath.Abs(globalFlags.OpenStackPrivateKey)
		if err != nil {
			return storage.State{}, err
		}

		_, key, err := c.readKey(absKeyPath)
		if err != nil {
			return storage.State{}, err
		}

		state.OpenStack.PrivateKey = key
	}

	return state, nil
}

func (c Config) updateVSphereState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	copyFlagToState(globalFlags.VSphereVCenterUser, &state.VSphere.VCenterUser)
	copyFlagToState(globalFlags.VSphereVCenterPassword, &state.VSphere.VCenterPassword)
	copyFlagToState(globalFlags.VSphereVCenterIP, &state.VSphere.VCenterIP)
	copyFlagToState(globalFlags.VSphereVCenterDC, &state.VSphere.VCenterDC)
	copyFlagToState(globalFlags.VSphereVCenterRP, &state.VSphere.VCenterRP)
	copyFlagToState(globalFlags.VSphereCluster, &state.VSphere.Cluster)
	copyFlagToState(globalFlags.VSphereNetwork, &state.VSphere.Network)
	copyFlagToState(globalFlags.VSphereVCenterDS, &state.VSphere.VCenterDS)
	copyFlagToState(globalFlags.VSphereSubnet, &state.VSphere.Subnet)

	return state, nil
}

func (c Config) updateAWSState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	copyFlagToState(globalFlags.AWSAccessKeyID, &state.AWS.AccessKeyID)
	copyFlagToState(globalFlags.AWSSecretAccessKey, &state.AWS.SecretAccessKey)

	if globalFlags.AWSRegion != "" {
		if state.AWS.Region != "" && globalFlags.AWSRegion != state.AWS.Region {
			return storage.State{}, fmt.Errorf("The region cannot be changed for an existing environment. The current region is %s.", state.AWS.Region)
		}
		state.AWS.Region = globalFlags.AWSRegion
	}

	return state, nil
}

func (c Config) updateAzureState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	copyFlagToState(globalFlags.AzureClientID, &state.Azure.ClientID)
	copyFlagToState(globalFlags.AzureClientSecret, &state.Azure.ClientSecret)
	copyFlagToState(globalFlags.AzureRegion, &state.Azure.Region)
	copyFlagToState(globalFlags.AzureSubscriptionID, &state.Azure.SubscriptionID)
	copyFlagToState(globalFlags.AzureTenantID, &state.Azure.TenantID)

	return state, nil
}

func (c Config) updateGCPState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	if globalFlags.GCPServiceAccountKey != "" {
		path, key, err := c.getGCPServiceAccountKey(globalFlags.GCPServiceAccountKey)
		if err != nil {
			return storage.State{}, err
		}
		state.GCP.ServiceAccountKey = key
		state.GCP.ServiceAccountKeyPath = path

		id, err := c.getGCPProjectID(key)
		if err != nil {
			return storage.State{}, err
		}
		if state.GCP.ProjectID != "" && id != state.GCP.ProjectID {
			return storage.State{}, fmt.Errorf("The project ID cannot be changed for an existing environment. The current project ID is %s.", state.GCP.ProjectID)
		}
		state.GCP.ProjectID = id
	}

	if globalFlags.GCPRegion != "" {
		if state.GCP.Region != "" && globalFlags.GCPRegion != state.GCP.Region {
			return storage.State{}, fmt.Errorf("The region cannot be changed for an existing environment. The current region is %s.", state.GCP.Region)
		}
		state.GCP.Region = globalFlags.GCPRegion
	}

	return state, nil
}

func (c Config) getGCPServiceAccountKey(key string) (string, string, error) {
	if _, err := c.fs.Stat(key); err != nil {
		return c.writeGCPServiceAccountKey(key)
	}
	return c.readKey(key)
}

func (c Config) writeGCPServiceAccountKey(contents string) (string, string, error) {
	tempFile, err := c.fs.TempFile("", "gcpServiceAccountKey.json")
	if err != nil {
		return "", "", fmt.Errorf("Creating temp file for credentials: %s", err)
	}
	err = c.fs.WriteFile(tempFile.Name(), []byte(contents), storage.StateMode)
	if err != nil {
		return "", "", fmt.Errorf("Writing credentials to temp file: %s", err)
	}
	return tempFile.Name(), contents, nil
}

func (c Config) readKey(path string) (string, string, error) {
	keyBytes, err := c.fs.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("Reading key: %v", err)
	}
	return path, string(keyBytes), nil
}

func (c Config) getGCPProjectID(key string) (string, error) {
	p := struct {
		ProjectID string `json:"project_id"`
	}{}

	err := json.Unmarshal([]byte(key), &p)
	if err != nil {
		return "", fmt.Errorf("Unmarshalling service account key (must be valid json): %s", err)
	}

	if p.ProjectID == "" {
		return "", errors.New("Service account key is missing field `project_id`")
	}

	return p.ProjectID, nil
}

func ValidateIAAS(state storage.State) error {
	var err error
	switch state.IAAS {
	case "aws":
		creds := []string{state.AWS.AccessKeyID, state.AWS.SecretAccessKey, state.AWS.Region}
		err = validate("AWS", creds)

	case "azure":
		creds := []string{state.Azure.ClientID, state.Azure.ClientSecret, state.Azure.Region, state.Azure.SubscriptionID, state.Azure.TenantID}
		err = validate("Azure", creds)

	case "gcp":
		creds := []string{state.GCP.ServiceAccountKey, state.GCP.Region}
		err = validate("GCP", creds)

	case "vsphere":
		creds := []string{state.VSphere.VCenterUser, state.VSphere.VCenterPassword, state.VSphere.VCenterIP, state.VSphere.VCenterIP,
			state.VSphere.VCenterDC, state.VSphere.Cluster, state.VSphere.VCenterRP, state.VSphere.Network, state.VSphere.VCenterDS, state.VSphere.Subnet}
		err = validate("vSphere", creds)

	case "openstack":
		creds := []string{state.OpenStack.InternalCidr, state.OpenStack.ExternalIP, state.OpenStack.AuthURL, state.OpenStack.AZ, state.OpenStack.DefaultKeyName,
			state.OpenStack.DefaultSecurityGroup, state.OpenStack.NetworkID, state.OpenStack.Password, state.OpenStack.Username, state.OpenStack.Project,
			state.OpenStack.Domain, state.OpenStack.Region, state.OpenStack.PrivateKey}
		err = validate("OpenStack", creds)

	default:
		err = errors.New("--iaas [gcp, aws, azure, vsphere, openstack] must be provided or BBL_IAAS must be set")
	}

	if err != nil {
		return fmt.Errorf("\n\n%s\n", err)
	}

	return nil
}

func NeedsIAASCreds(command string) bool {
	_, ok := map[string]struct{}{
		"up":                struct{}{},
		"down":              struct{}{},
		"plan":              struct{}{},
		"destroy":           struct{}{},
		"leftovers":         struct{}{},
		"cleanup-leftovers": struct{}{},
		"rotate":            struct{}{},
	}[command]
	return ok
}

func validate(iaas string, creds []string) error {
	for _, s := range creds {
		if s == "" {
			return fmt.Errorf("There are %s credentials missing. To see all required credentials run `bbl plan --help`.", iaas)
		}
	}
	return nil
}
