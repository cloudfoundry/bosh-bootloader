package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	flags "github.com/jessevdk/go-flags"
)

type globalFlags struct {
	Help     bool   `short:"h" long:"help"`
	Debug    bool   `short:"d" long:"debug"         env:"BBL_DEBUG"`
	Version  bool   `short:"v" long:"version"`
	StateDir string `short:"s" long:"state-dir"     env:"BBL_STATE_DIR"`
	IAAS     string `long:"iaas"                    env:"BBL_IAAS"`

	AWSAccessKeyID     string `long:"aws-access-key-id"       env:"BBL_AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `long:"aws-secret-access-key"   env:"BBL_AWS_SECRET_ACCESS_KEY"`
	AWSRegion          string `long:"aws-region"              env:"BBL_AWS_REGION"`

	AzureSubscriptionID string `long:"azure-subscription-id"  env:"BBL_AZURE_SUBSCRIPTION_ID"`
	AzureTenantID       string `long:"azure-tenant-id"        env:"BBL_AZURE_TENANT_ID"`
	AzureClientID       string `long:"azure-client-id"        env:"BBL_AZURE_CLIENT_ID"`
	AzureClientSecret   string `long:"azure-client-secret"    env:"BBL_AZURE_CLIENT_SECRET"`

	GCPServiceAccountKey string `long:"gcp-service-account-key" env:"BBL_GCP_SERVICE_ACCOUNT_KEY"`
	GCPProjectID         string `long:"gcp-project-id"          env:"BBL_GCP_PROJECT_ID"`
	GCPZone              string `long:"gcp-zone"                env:"BBL_GCP_ZONE"`
	GCPRegion            string `long:"gcp-region"              env:"BBL_GCP_REGION"`
}

type ParsedFlags struct {
	State         storage.State
	RemainingArgs []string
	Help          bool
	Debug         bool
	Version       bool
	StateDir      string
}

func NewConfig(getState func(string) (storage.State, error)) Config {
	return Config{
		getState: getState,
	}
}

type Config struct {
	getState func(string) (storage.State, error)
}

func (c Config) Bootstrap(args []string) (ParsedFlags, error) {
	var globalFlags globalFlags

	parser := flags.NewParser(&globalFlags, flags.IgnoreUnknown)

	remainingArgs, err := parser.ParseArgs(args[1:])
	if err != nil {
		return ParsedFlags{}, err
	}

	stateDir := globalFlags.StateDir
	if stateDir == "" {
		stateDir, err = os.Getwd()
		if err != nil {
			// not tested
			return ParsedFlags{}, err
		}
	}

	state, err := c.getState(stateDir)
	if err != nil {
		return ParsedFlags{}, err
	}

	if globalFlags.IAAS != "" {
		if state.IAAS != "" && globalFlags.IAAS != state.IAAS {
			iaasMismatch := fmt.Sprintf("The iaas type cannot be changed for an existing environment. The current iaas type is %s.", state.IAAS)
			return ParsedFlags{}, errors.New(iaasMismatch)
		}
		state.IAAS = globalFlags.IAAS
	}

	if globalFlags.AWSAccessKeyID != "" {
		state.AWS.AccessKeyID = globalFlags.AWSAccessKeyID
	}
	if globalFlags.AWSSecretAccessKey != "" {
		state.AWS.SecretAccessKey = globalFlags.AWSSecretAccessKey
	}
	if globalFlags.AWSRegion != "" {
		if state.AWS.Region != "" && globalFlags.AWSRegion != state.AWS.Region {
			regionMismatch := fmt.Sprintf("The region cannot be changed for an existing environment. The current region is %s.", state.AWS.Region)
			return ParsedFlags{}, errors.New(regionMismatch)
		}
		state.AWS.Region = globalFlags.AWSRegion
	}

	if globalFlags.GCPServiceAccountKey != "" {
		serviceAccountKey, err := parseServiceAccountKey(globalFlags.GCPServiceAccountKey)
		if err != nil {
			return ParsedFlags{}, err
		}
		state.GCP.ServiceAccountKey = serviceAccountKey
	}
	if globalFlags.GCPProjectID != "" {
		state.GCP.ProjectID = globalFlags.GCPProjectID
	}
	if globalFlags.GCPZone != "" {
		state.GCP.Zone = globalFlags.GCPZone
	}
	if globalFlags.GCPRegion != "" {
		if state.GCP.Region != "" && globalFlags.GCPRegion != state.GCP.Region {
			regionMismatch := fmt.Sprintf("The region cannot be changed for an existing environment. The current region is %s.", state.GCP.Region)
			return ParsedFlags{}, errors.New(regionMismatch)
		}
		state.GCP.Region = globalFlags.GCPRegion
	}
	if globalFlags.AzureSubscriptionID != "" {
		state.Azure.SubscriptionID = globalFlags.AzureSubscriptionID
	}
	if globalFlags.AzureTenantID != "" {
		state.Azure.TenantID = globalFlags.AzureTenantID
	}
	if globalFlags.AzureClientID != "" {
		state.Azure.ClientID = globalFlags.AzureClientID
	}
	if globalFlags.AzureClientSecret != "" {
		state.Azure.ClientSecret = globalFlags.AzureClientSecret
	}

	nonStatefulCommand := len(remainingArgs) == 0 || (remainingArgs[0] == "help" || remainingArgs[0] == "version" || remainingArgs[0] == "latest-error")
	ignoreMissingFlags := globalFlags.Help || globalFlags.Version || nonStatefulCommand

	if !ignoreMissingFlags {
		err := validate(state)
		if err != nil {
			return ParsedFlags{}, err
		}
	}

	return ParsedFlags{State: state, RemainingArgs: remainingArgs, Help: globalFlags.Help, Debug: globalFlags.Debug, Version: globalFlags.Version, StateDir: globalFlags.StateDir}, nil
}

func validate(state storage.State) error {
	if state.IAAS == "" || (state.IAAS != "gcp" && state.IAAS != "aws" && state.IAAS != "azure") {
		return errors.New("--iaas [gcp, aws, azure] must be provided or BBL_IAAS must be set")
	}
	if state.IAAS == "aws" {
		err := validateAWSFlags(state.AWS)
		if err != nil {
			return err
		}
	}
	if state.IAAS == "gcp" {
		err := validateGCPFlags(state.GCP)
		if err != nil {
			return err
		}
	}
	if state.IAAS == "azure" {
		err := validateAzureFlags(state.Azure)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateAWSFlags(awsFlags storage.AWS) error {
	if awsFlags.AccessKeyID == "" {
		return errors.New("AWS access key ID must be provided")
	}
	if awsFlags.SecretAccessKey == "" {
		return errors.New("AWS secret access key must be provided")
	}
	if awsFlags.Region == "" {
		return errors.New("AWS region must be provided")
	}
	return nil
}

func validateGCPFlags(gcpFlags storage.GCP) error {
	if gcpFlags.ServiceAccountKey == "" {
		return errors.New("GCP service account key must be provided")
	}
	if gcpFlags.ProjectID == "" {
		return errors.New("GCP project ID must be provided")
	}
	if gcpFlags.Zone == "" {
		return errors.New("GCP zone must be provided")
	}
	if gcpFlags.Region == "" {
		return errors.New("GCP region must be provided")
	}
	return nil
}

func validateAzureFlags(azureFlags storage.Azure) error {
	if azureFlags.SubscriptionID == "" {
		return errors.New("Azure subscription id must be provided")
	}
	if azureFlags.TenantID == "" {
		return errors.New("Azure tenant id must be provided")
	}
	if azureFlags.ClientID == "" {
		return errors.New("Azure client id must be provided")
	}
	if azureFlags.ClientSecret == "" {
		return errors.New("Azure client secret must be provided")
	}
	return nil
}

func parseServiceAccountKey(serviceAccountKey string) (string, error) {
	var key string

	if _, err := os.Stat(serviceAccountKey); err != nil {
		key = serviceAccountKey
	} else {
		rawServiceAccountKey, err := ioutil.ReadFile(serviceAccountKey)
		if err != nil {
			return "", fmt.Errorf("error reading service account key from file: %v", err)
		}

		key = string(rawServiceAccountKey)
	}

	var tmp interface{}
	err := json.Unmarshal([]byte(key), &tmp)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling service account key (must be valid json): %v", err)
	}

	return key, err
}
