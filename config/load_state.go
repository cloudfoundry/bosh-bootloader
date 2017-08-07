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

type GlobalFlags struct {
	Help                 bool   `short:"h" long:"help"`
	Debug                bool   `short:"d" long:"debug"         env:"BBL_DEBUG"`
	Version              bool   `short:"v" long:"version"`
	StateDir             string `short:"s" long:"state-dir"     env:"BBL_STATE_DIR"`
	IAAS                 string `long:"iaas"                    env:"BBL_IAAS"`
	AWSAccessKeyID       string `long:"aws-access-key-id"       env:"BBL_AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey   string `long:"aws-secret-access-key"   env:"BBL_AWS_SECRET_ACCESS_KEY"`
	AWSRegion            string `long:"aws-region"              env:"BBL_AWS_REGION"`
	GCPServiceAccountKey string `long:"gcp-service-account-key" env:"BBL_GCP_SERVICE_ACCOUNT_KEY"`
	GCPProjectID         string `long:"gcp-project-id"          env:"BBL_GCP_PROJECT_ID"`
	GCPZone              string `long:"gcp-zone"                env:"BBL_GCP_ZONE"`
	GCPRegion            string `long:"gcp-region"              env:"BBL_GCP_REGION"`
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
	var globalFlags GlobalFlags

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

	if state.IAAS == "" || (state.IAAS != "gcp" && state.IAAS != "aws") {
		return ParsedFlags{}, errors.New("--iaas [gcp, aws] must be provided or BBL_IAAS must be set")
	}
	if state.IAAS == "aws" {
		if state.AWS.AccessKeyID == "" {
			return ParsedFlags{}, errors.New("AWS access key ID must be provided")
		}
		if state.AWS.SecretAccessKey == "" {
			return ParsedFlags{}, errors.New("AWS secret access key must be provided")
		}
		if state.AWS.Region == "" {
			return ParsedFlags{}, errors.New("AWS region must be provided")
		}
	}
	if state.IAAS == "gcp" {
		if state.GCP.ServiceAccountKey == "" {
			return ParsedFlags{}, errors.New("GCP service account key must be provided")
		}
		if state.GCP.ProjectID == "" {
			return ParsedFlags{}, errors.New("GCP project ID must be provided")
		}
		if state.GCP.Zone == "" {
			return ParsedFlags{}, errors.New("GCP zone must be provided")
		}
		if state.GCP.Region == "" {
			return ParsedFlags{}, errors.New("GCP region must be provided")
		}
	}

	return ParsedFlags{State: state, RemainingArgs: remainingArgs, Help: globalFlags.Help, Debug: globalFlags.Debug, Version: globalFlags.Version, StateDir: globalFlags.StateDir}, nil
}

type ParsedFlags struct {
	State         storage.State
	RemainingArgs []string
	Help          bool
	Debug         bool
	Version       bool
	StateDir      string
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
