package integration

import (
	"errors"
	"fmt"
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
	StemcellName             string
	StemcellPath             string
	GardenReleasePath        string
	ConcourseReleasePath     string
}

func LoadConfig() (Config, error) {
	config := loadConfigFromEnvVars()

	awsCredsValidationError := validateAWSCreds(config)
	gcpCredsValidationError := validateGCPCreds(config)

	if awsCredsValidationError == nil && gcpCredsValidationError == nil {
		return Config{}, errors.New("Multiple IAAS Credentials provided:\n Provide a set of credentials for a single IAAS.")
	}

	if awsCredsValidationError != nil && gcpCredsValidationError != nil {
		return Config{}, fmt.Errorf("Multiple Credential Errors Found: %s\n%s\nProvide a full set of credentials for a single IAAS.", awsCredsValidationError, gcpCredsValidationError)
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

func validateAWSCreds(config Config) error {
	if config.AWSAccessKeyID == "" {
		return errors.New("aws access key id is missing")
	}

	if config.AWSSecretAccessKey == "" {
		return errors.New("aws secret access key is missing")
	}

	if config.AWSRegion == "" {
		return errors.New("aws region is missing")
	}

	return nil
}

func validateGCPCreds(config Config) error {
	if config.GCPServiceAccountKeyPath == "" {
		return errors.New("gcp service account key path is missing")
	}

	if config.GCPProjectID == "" {
		return errors.New("project id is missing")
	}

	if config.GCPRegion == "" {
		return errors.New("gcp region is missing")
	}

	if config.GCPZone == "" {
		return errors.New("gcp zone is missing")
	}

	return nil
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
		StemcellName:             os.Getenv("STEMCELL_NAME"),
		StemcellPath:             os.Getenv("STEMCELL_PATH"),
		GardenReleasePath:        os.Getenv("GARDEN_RELEASE_PATH"),
		ConcourseReleasePath:     os.Getenv("CONCOURSE_RELEASE_PATH"),
	}
}
