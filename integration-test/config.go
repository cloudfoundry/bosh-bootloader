package integration

import (
	"errors"
	"io/ioutil"
	"os"
)

type Config struct {
	AWSAccessKeyID           string
	AWSSecretAccessKey       string
	AWSRegion                string
	GCPServiceAccountKeyPath string
	GCPProjectID             string
	GCPRegion                string
	GCPZone                  string
	GCPEnvPrefix             string
	StateFileDir             string
}

func LoadAWSConfig() (Config, error) {
	config := loadConfigFromEnvVars()

	if config.AWSAccessKeyID == "" {
		return Config{}, errors.New("aws access key id is missing")
	}

	if config.AWSSecretAccessKey == "" {
		return Config{}, errors.New("aws secret access key is missing")
	}

	if config.AWSRegion == "" {
		return Config{}, errors.New("aws region is missing")
	}

	if config.StateFileDir == "" {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			return Config{}, err
		}
		config.StateFileDir = dir
	}

	return config, nil
}

func LoadGCPConfig() (Config, error) {
	config := loadConfigFromEnvVars()

	if config.GCPServiceAccountKeyPath == "" {
		return Config{}, errors.New("gcp service account key path is missing")
	}

	if config.GCPProjectID == "" {
		return Config{}, errors.New("project id is missing")
	}

	if config.GCPRegion == "" {
		return Config{}, errors.New("gcp region is missing")
	}

	if config.GCPZone == "" {
		return Config{}, errors.New("gcp zone is missing")
	}

	if config.StateFileDir == "" {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			return Config{}, err
		}
		config.StateFileDir = dir
	}

	return config, nil
}

func loadConfigFromEnvVars() Config {
	return Config{
		AWSAccessKeyID:           os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey:       os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:                os.Getenv("AWS_REGION"),
		GCPServiceAccountKeyPath: os.Getenv("GCP_SERVICE_ACCOUNT_KEY"),
		GCPProjectID:             os.Getenv("GCP_PROJECT_ID"),
		GCPRegion:                os.Getenv("GCP_REGION"),
		GCPZone:                  os.Getenv("GCP_ZONE"),
		GCPEnvPrefix:             os.Getenv("GCP_ENV_PREFIX"),
		StateFileDir:             os.Getenv("STATE_DIR"),
	}
}
