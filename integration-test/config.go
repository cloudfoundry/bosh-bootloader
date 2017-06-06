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
	StemcellPath             string
	GardenReleasePath        string
	ConcourseReleasePath     string
	ConcourseDeploymentPath  string
	EnableTerraformFlag      bool
}

func LoadConfig() (Config, error) {
	config := loadConfigFromEnvVars()

	awsCredsValidationError := validateAWSCreds(config)
	gcpCredsValidationError := validateGCPCreds(config)

	if awsCredsValidationError == nil && gcpCredsValidationError == nil {
		return Config{}, errors.New("Multiple IAAS Credentials provided:\n Provide a set of credentials for a single IAAS.")
	}

	if awsCredsValidationError != nil && gcpCredsValidationError != nil {
		return Config{}, fmt.Errorf("Multiple Credential Errors Found: %s\nProvide a full set of credentials for a single IAAS.", gcpCredsValidationError)
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
		GCPServiceAccountKeyPath: os.Getenv("BBL_GCP_SERVICE_ACCOUNT_KEY"),
		GCPProjectID:             os.Getenv("BBL_GCP_PROJECT_ID"),
		GCPRegion:                os.Getenv("BBL_GCP_REGION"),
		GCPZone:                  os.Getenv("BBL_GCP_ZONE"),
		GCPEnvPrefix:             os.Getenv("BBL_GCP_ENV_PREFIX"),
		StateFileDir:             os.Getenv("BBL_STATE_DIR"),
		StemcellPath:             os.Getenv("STEMCELL_PATH"),
		GardenReleasePath:        os.Getenv("GARDEN_RELEASE_PATH"),
		ConcourseReleasePath:     os.Getenv("CONCOURSE_RELEASE_PATH"),
		ConcourseDeploymentPath:  os.Getenv("CONCOURSE_DEPLOYMENT_PATH"),
		EnableTerraformFlag:      os.Getenv("ENABLE_TERRAFORM_FLAG") == "true",
	}
}
