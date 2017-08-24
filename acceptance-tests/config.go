package acceptance

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	IAAS string

	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string

	AzureSubscriptionID string
	AzureTenantID       string
	AzureClientID       string
	AzureClientSecret   string

	GCPServiceAccountKey string
	GCPProjectID         string
	GCPRegion            string
	GCPZone              string

	StateFileDir            string
	StemcellPath            string
	GardenReleasePath       string
	ConcourseReleasePath    string
	ConcourseDeploymentPath string
}

func LoadConfig() (Config, error) {
	config := loadConfigFromEnvVars()

	err := validateIAAS(config)
	if err != nil {
		return Config{}, fmt.Errorf("Error found: %s\n", err)
	}

	switch config.IAAS {
	case "aws":
		err = validateAWSCreds(config)
		if err != nil {
			return Config{}, fmt.Errorf("Error Found: %s\nProvide a full set of credentials for a single IAAS.", err)
		}
	case "gcp":
		err = validateGCPCreds(config)
		if err != nil {
			return Config{}, fmt.Errorf("Error Found: %s\nProvide a full set of credentials for a single IAAS.", err)
		}
	case "azure":
		err = validateAzureCreds(config)
		if err != nil {
			return Config{}, fmt.Errorf("Error Found: %s\nProvide a full set of credentials for a single IAAS.", err)
		}
	}

	if config.StateFileDir == "" {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			return Config{}, err
		}
		config.StateFileDir = dir
	}
	fmt.Printf("using state-dir: %s\n", config.StateFileDir)

	return config, nil
}

func validateIAAS(config Config) error {
	if config.IAAS == "" {
		return errors.New("iaas is missing")
	}

	return nil
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

func validateAzureCreds(config Config) error {
	if config.AzureSubscriptionID == "" {
		return errors.New("azure subscription id is missing")
	}

	if config.AzureTenantID == "" {
		return errors.New("azure tenant id is missing")
	}

	if config.AzureClientID == "" {
		return errors.New("azure client id is missing")
	}

	if config.AzureClientSecret == "" {
		return errors.New("azure client secret is missing")
	}

	return nil
}

func validateGCPCreds(config Config) error {
	if config.GCPServiceAccountKey == "" {
		return errors.New("gcp service account key is missing")
	}

	if config.GCPProjectID == "" {
		return errors.New("gcp project id is missing")
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
		IAAS: os.Getenv("BBL_IAAS"),

		AWSAccessKeyID:     os.Getenv("BBL_AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("BBL_AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          os.Getenv("BBL_AWS_REGION"),

		AzureSubscriptionID: os.Getenv("BBL_AZURE_SUBSCRIPTION_ID"),
		AzureTenantID:       os.Getenv("BBL_AZURE_TENANT_ID"),
		AzureClientID:       os.Getenv("BBL_AZURE_CLIENT_ID"),
		AzureClientSecret:   os.Getenv("BBL_AZURE_CLIENT_SECRET"),

		GCPServiceAccountKey: os.Getenv("BBL_GCP_SERVICE_ACCOUNT_KEY"),
		GCPProjectID:         os.Getenv("BBL_GCP_PROJECT_ID"),
		GCPRegion:            os.Getenv("BBL_GCP_REGION"),
		GCPZone:              os.Getenv("BBL_GCP_ZONE"),

		StateFileDir:            os.Getenv("BBL_STATE_DIR"),
		StemcellPath:            os.Getenv("STEMCELL_PATH"),
		GardenReleasePath:       os.Getenv("GARDEN_RELEASE_PATH"),
		ConcourseReleasePath:    os.Getenv("CONCOURSE_RELEASE_PATH"),
		ConcourseDeploymentPath: os.Getenv("CONCOURSE_DEPLOYMENT_PATH"),
	}
}
