package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	flags "github.com/jessevdk/go-flags"
)

type globalFlags struct {
	Help     bool   `short:"h" long:"help"`
	Debug    bool   `short:"d" long:"debug"     env:"BBL_DEBUG"`
	Version  bool   `short:"v" long:"version"`
	StateDir string `short:"s" long:"state-dir"`
	IAAS     string `          long:"iaas"      env:"BBL_IAAS"`

	AWSAccessKeyID     string `long:"aws-access-key-id"       env:"BBL_AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `long:"aws-secret-access-key"   env:"BBL_AWS_SECRET_ACCESS_KEY"`
	AWSRegion          string `long:"aws-region"              env:"BBL_AWS_REGION"`

	AzureClientID       string `long:"azure-client-id"        env:"BBL_AZURE_CLIENT_ID"`
	AzureClientSecret   string `long:"azure-client-secret"    env:"BBL_AZURE_CLIENT_SECRET"`
	AzureRegion         string `long:"azure-region"           env:"BBL_AZURE_REGION"`
	AzureSubscriptionID string `long:"azure-subscription-id"  env:"BBL_AZURE_SUBSCRIPTION_ID"`
	AzureTenantID       string `long:"azure-tenant-id"        env:"BBL_AZURE_TENANT_ID"`

	GCPServiceAccountKey string `long:"gcp-service-account-key" env:"BBL_GCP_SERVICE_ACCOUNT_KEY"`
	GCPProjectID         string `long:"gcp-project-id"          env:"BBL_GCP_PROJECT_ID"`
	GCPZone              string `long:"gcp-zone"                env:"BBL_GCP_ZONE"`
	GCPRegion            string `long:"gcp-region"              env:"BBL_GCP_REGION"`

	VSphereCluster         string `long:"vsphere-vcenter-cluster"  env:"BBL_VSPHERE_VCENTER_CLUSTER"`
	VSphereNetwork         string `long:"vsphere-network"          env:"BBL_VSPHERE_NETWORK"`
	VSphereSubnet          string `long:"vsphere-subnet"           env:"BBL_VSPHERE_SUBNET"`
	VSphereVCenterDC       string `long:"vsphere-vcenter-dc"       env:"BBL_VSPHERE_VCENTER_DC"`
	VSphereVCenterDS       string `long:"vsphere-vcenter-ds"       env:"BBL_VSPHERE_VCENTER_DS"`
	VSphereVCenterIP       string `long:"vsphere-vcenter-ip"       env:"BBL_VSPHERE_VCENTER_IP"`
	VSphereVCenterPassword string `long:"vsphere-vcenter-password" env:"BBL_VSPHERE_VCENTER_PASSWORD"`
	VSphereVCenterRP       string `long:"vsphere-vcenter-rp"       env:"BBL_VSPHERE_VCENTER_RP"`
	VSphereVCenterUser     string `long:"vsphere-vcenter-user"     env:"BBL_VSPHERE_VCENTER_USER"`
}

type logger interface {
	Println(string)
}

type StateBootstrap interface {
	GetState(string) (storage.State, error)
}

type migrator interface {
	Migrate(storage.State) (storage.State, error)
}

func NewConfig(bootstrap StateBootstrap, migrator migrator, logger logger) Config {
	return Config{
		stateBootstrap: bootstrap,
		migrator:       migrator,
		logger:         logger,
	}
}

type Config struct {
	stateBootstrap StateBootstrap
	migrator       migrator
	logger         logger
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

	if globalFlags.GCPProjectID != "" {
		c.logger.Println("Deprecation warning: the --gcp-project-id flag (BBL_GCP_PROJECT_ID) is now ignored.")
	}

	if globalFlags.GCPZone != "" {
		c.logger.Println("Deprecation warning: the --gcp-zone flag (BBL_GCP_ZONE) is now ignored.")
	}

	if command == "bosh-deployment-vars" {
		c.logger.Println(`Deprecation warning: the bosh-deployment-vars command has been deprecated and will be removed in bbl v6.0.0. The bosh deployment vars are stored in the vars directory.`)
	}

	if command == "jumpbox-deployment-vars" {
		c.logger.Println(`Deprecation warning: the jumpbox-deployment-vars command has been deprecated and will be removed in bbl v6.0.0. The jumpbox deployment vars are stored in the vars directory.`)
	}

	if command == "create-lbs" {
		c.logger.Println(`Deprecation warning: the create-lbs command has been deprecated and will be removed in bbl v6.0.0. Create load balancers with "plan" or "up" e.g. "bbl up --lb-type <type> --lb-cert <cert> --lb-key <key>" or "bbl up --lb-type <type> --lb-cert <cert> --lb-key <key>".`)
	}

	if command == "delete-lbs" {
		c.logger.Println(`Deprecation warning: the delete-lbs command has been deprecated and will be removed in bbl v6.0.0. Delete load balancers by calling "plan" without the lb flags.`)
	}

	state, err := c.stateBootstrap.GetState(globalFlags.StateDir)
	if err != nil {
		return application.Configuration{}, err
	}

	state, err = c.migrator.Migrate(state)
	if err != nil {
		return application.Configuration{}, err
	}

	state, err = updateIAASState(globalFlags, state)
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

func updateIAASState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	if globalFlags.IAAS != "" {
		if state.IAAS != "" && globalFlags.IAAS != state.IAAS {
			iaasMismatch := fmt.Sprintf("The iaas type cannot be changed for an existing environment. The current iaas type is %s.", state.IAAS)
			return storage.State{}, errors.New(iaasMismatch)
		}
		state.IAAS = globalFlags.IAAS
	}

	switch state.IAAS {
	case "aws":
		state, err := updateAWSState(globalFlags, state)
		return state, err
	case "azure":
		state, err := updateAzureState(globalFlags, state)
		return state, err
	case "gcp":
		state, err := updateGCPState(globalFlags, state)
		return state, err
	case "vsphere":
		state, err := updateVSphereState(globalFlags, state)
		return state, err
	}

	return state, nil
}

func updateAWSState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	if globalFlags.AWSAccessKeyID != "" {
		state.AWS.AccessKeyID = globalFlags.AWSAccessKeyID
	}
	if globalFlags.AWSSecretAccessKey != "" {
		state.AWS.SecretAccessKey = globalFlags.AWSSecretAccessKey
	}
	if globalFlags.AWSRegion != "" {
		if state.AWS.Region != "" && globalFlags.AWSRegion != state.AWS.Region {
			regionMismatch := fmt.Sprintf("The region cannot be changed for an existing environment. The current region is %s.", state.AWS.Region)
			return storage.State{}, errors.New(regionMismatch)
		}
		state.AWS.Region = globalFlags.AWSRegion
	}

	return state, nil
}

func updateAzureState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	if globalFlags.AzureClientID != "" {
		state.Azure.ClientID = globalFlags.AzureClientID
	}
	if globalFlags.AzureClientSecret != "" {
		state.Azure.ClientSecret = globalFlags.AzureClientSecret
	}
	if globalFlags.AzureRegion != "" {
		state.Azure.Region = globalFlags.AzureRegion
	}
	if globalFlags.AzureSubscriptionID != "" {
		state.Azure.SubscriptionID = globalFlags.AzureSubscriptionID
	}
	if globalFlags.AzureTenantID != "" {
		state.Azure.TenantID = globalFlags.AzureTenantID
	}

	return state, nil
}

func updateGCPState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	if globalFlags.GCPServiceAccountKey != "" {
		serviceAccountKey, projectID, err := parseServiceAccountKey(globalFlags.GCPServiceAccountKey)
		if err != nil {
			return storage.State{}, err
		}
		state.GCP.ServiceAccountKey = serviceAccountKey
		if state.GCP.ProjectID != "" && projectID != state.GCP.ProjectID {
			projectIDMismatch := fmt.Sprintf("The project ID cannot be changed for an existing environment. The current project ID is %s.", state.GCP.ProjectID)
			return storage.State{}, errors.New(projectIDMismatch)
		}
		state.GCP.ProjectID = projectID
	}
	if globalFlags.GCPRegion != "" {
		if state.GCP.Region != "" && globalFlags.GCPRegion != state.GCP.Region {
			regionMismatch := fmt.Sprintf("The region cannot be changed for an existing environment. The current region is %s.", state.GCP.Region)
			return storage.State{}, errors.New(regionMismatch)
		}
		state.GCP.Region = globalFlags.GCPRegion
	}

	return state, nil
}

func updateVSphereState(globalFlags globalFlags, state storage.State) (storage.State, error) {
	if globalFlags.VSphereVCenterUser != "" {
		state.VSphere.VCenterUser = globalFlags.VSphereVCenterUser
	}
	if globalFlags.VSphereVCenterPassword != "" {
		state.VSphere.VCenterPassword = globalFlags.VSphereVCenterPassword
	}
	if globalFlags.VSphereVCenterIP != "" {
		state.VSphere.VCenterIP = globalFlags.VSphereVCenterIP
	}
	if globalFlags.VSphereVCenterDC != "" {
		state.VSphere.VCenterDC = globalFlags.VSphereVCenterDC
	}
	if globalFlags.VSphereCluster != "" {
		state.VSphere.Cluster = globalFlags.VSphereCluster
	}
	if globalFlags.VSphereVCenterRP != "" {
		state.VSphere.VCenterRP = globalFlags.VSphereVCenterRP
	}
	if globalFlags.VSphereNetwork != "" {
		state.VSphere.Network = globalFlags.VSphereNetwork
	}
	if globalFlags.VSphereVCenterDS != "" {
		state.VSphere.VCenterDS = globalFlags.VSphereVCenterDS
	}
	if globalFlags.VSphereSubnet != "" {
		state.VSphere.Subnet = globalFlags.VSphereSubnet
	}

	return state, nil
}

func supportedIAAS(iaas string) bool {
	supported := []string{"aws", "azure", "gcp", "vsphere"}
	for _, i := range supported {
		if iaas == i {
			return true
		}
	}
	return false
}

func ValidateIAAS(state storage.State) error {
	if !supportedIAAS(state.IAAS) {
		return errors.New("--iaas [gcp, aws, azure] must be provided or BBL_IAAS must be set")
	}
	if state.IAAS == "aws" {
		err := validateAWS(state.AWS)
		if err != nil {
			return err
		}
	}
	if state.IAAS == "azure" {
		err := validateAzure(state.Azure)
		if err != nil {
			return err
		}
	}
	if state.IAAS == "gcp" {
		err := validateGCP(state.GCP)
		if err != nil {
			return err
		}
	}
	if state.IAAS == "vsphere" {
		err := validateVSphere(state.VSphere)
		if err != nil {
			return err
		}
	}
	return nil
}

func NeedsIAASCreds(command string) bool {
	_, ok := map[string]struct{}{
		"up":         struct{}{},
		"down":       struct{}{},
		"plan":       struct{}{},
		"destroy":    struct{}{},
		"create-lbs": struct{}{},
		"delete-lbs": struct{}{},
		"update-lbs": struct{}{},
		"rotate":     struct{}{},
	}[command]
	return ok
}

func validateAWS(aws storage.AWS) error {
	if aws.AccessKeyID == "" {
		return errors.New("AWS access key ID must be provided (--aws-access-key-id or BBL_AWS_ACCESS_KEY_ID)")
	}
	if aws.SecretAccessKey == "" {
		return errors.New("AWS secret access key must be provided (--aws-secret-access-key or BBL_AWS_SECRET_ACCESS_KEY)")
	}
	if aws.Region == "" {
		return errors.New("AWS region must be provided (--aws-region or BBL_AWS_REGION)")
	}
	return nil
}

func validateAzure(azure storage.Azure) error {
	if azure.ClientID == "" {
		return errors.New("Azure client id must be provided (--azure-client-id or BBL_AZURE_CLIENT_ID)")
	}
	if azure.ClientSecret == "" {
		return errors.New("Azure client secret must be provided (--azure-client-secret or BBL_AZURE_CLIENT_SECRET)")
	}
	if azure.Region == "" {
		return errors.New("Azure region must be provided (--azure-region or BBL_AZURE_REGION)")
	}
	if azure.SubscriptionID == "" {
		return errors.New("Azure subscription id must be provided (--azure-subscription-id or BBL_AZURE_SUBSCRIPTION_ID)")
	}
	if azure.TenantID == "" {
		return errors.New("Azure tenant id must be provided (--azure-tenant-id or BBL_AZURE_TENANT_ID)")
	}
	return nil
}

func validateGCP(gcp storage.GCP) error {
	if gcp.ServiceAccountKey == "" {
		return errors.New("GCP service account key must be provided (--gcp-service-account-key or BBL_GCP_SERVICE_ACCOUNT_KEY)")
	}
	if gcp.Region == "" {
		return errors.New("GCP region must be provided (--gcp-region or BBL_GCP_REGION)")
	}
	return nil
}

func validateVSphere(vsphere storage.VSphere) error {
	if vsphere.VCenterUser == "" {
		return errors.New("vSphere vcenter user must be provided (--vsphere-vcenter-user or BBL_VSPHERE_VCENTER_USER)")
	}
	if vsphere.VCenterPassword == "" {
		return errors.New("vSphere vcenter password must be provided (--vsphere-vcenter-password or BBL_VSPHERE_VCENTER_PASSWORD)")
	}
	if vsphere.VCenterIP == "" {
		return errors.New("vSphere vcenter ip must be provided (--vsphere-vcenter-ip or BBL_VSPHERE_VCENTER_IP)")
	}
	if vsphere.VCenterDC == "" {
		return errors.New("vSphere vcenter datacenter must be provided (--vsphere-vcenter-dc or BBL_VSPHERE_VCENTER_DC)")
	}
	if vsphere.Cluster == "" {
		return errors.New("vSphere cluster must be provided (--vsphere-vcenter-cluster or BBL_VSPHERE_VCENTER_CLUSTER)")
	}
	if vsphere.VCenterRP == "" {
		return errors.New("vSphere vcenter resource pool must be provided (--vsphere-vcenter-rp or BBL_VSPHERE_VCENTER_RP)")
	}
	if vsphere.Network == "" {
		return errors.New("vSphere network must be provided (--vsphere-network or BBL_VSPHERE_NETWORK)")
	}
	if vsphere.VCenterDS == "" {
		return errors.New("vSphere vcenter datastore must be provided (--vsphere-vcenter-ds or BBL_VSPHERE_VCENTER_DS)")
	}
	if vsphere.Subnet == "" {
		return errors.New("vSphere subnet must be provided (--vsphere-subnet or BBL_VSPHERE_SUBNET)")
	}
	return nil
}

func parseServiceAccountKey(serviceAccountKey string) (string, string, error) {
	var key string

	if _, err := os.Stat(serviceAccountKey); err != nil {
		key = serviceAccountKey
	} else {
		rawServiceAccountKey, err := ioutil.ReadFile(serviceAccountKey)
		if err != nil {
			return "", "", fmt.Errorf("error reading service account key from file: %v", err)
		}

		key = string(rawServiceAccountKey)
	}

	p := struct {
		ProjectID string `json:"project_id"`
	}{}
	err := json.Unmarshal([]byte(key), &p)
	if err != nil {
		return "", "", fmt.Errorf("error unmarshalling service account key (must be valid json): %v", err)
	}
	if p.ProjectID == "" {
		return "", "", errors.New("service account key is missing field `project_id`")
	}

	return key, p.ProjectID, err
}
