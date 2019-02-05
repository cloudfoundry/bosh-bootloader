package acceptance

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
)

type Config struct {
	IAAS        string
	StemcellURL string

	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string

	AzureClientID       string
	AzureClientSecret   string
	AzureRegion         string
	AzureSubscriptionID string
	AzureTenantID       string

	GCPServiceAccountKey string
	GCPRegion            string

	VSphereNetwork          string
	VSphereSubnetCIDR       string
	VSphereVCenterIP        string
	VSphereVCenterUser      string
	VSphereVCenterPassword  string
	VSphereVCenterDC        string
	VSphereVCenterCluster   string
	VSphereVCenterRP        string
	VSphereVCenterDS        string
	VSphereVCenterDisks     string
	VSphereVCenterVMs       string
	VSphereVCenterTemplates string

	OpenStackAuthURL     string
	OpenStackAZ          string
	OpenStackNetworkID   string
	OpenStackNetworkName string
	OpenStackPassword    string
	OpenStackUsername    string
	OpenStackProject     string
	OpenStackDomain      string
	OpenStackRegion      string

	// if bucket is set, statefiledir should be ignored
	StateFileDir   string
	BBLStateBucket string
}

// Loads all of the credentials from environment variables
// to be used in the acceptance test. Validates the presence
// of required credentials before starting up tests.
func LoadConfig() (Config, error) {
	config := loadConfigFromEnvVars()

	err := validateIAAS(config)
	if err != nil {
		return Config{}, fmt.Errorf("Error found: %s\n", err)
	}

	switch config.IAAS {
	case "aws":
		err = validateAWSCreds(config)
	case "azure":
		err = validateAzureCreds(config)
	case "gcp":
		err = validateGCPCreds(config)
	case "vsphere":
		err = validateVSphereCreds(config)
	case "openstack":
		err = validateOpenStackCreds(config)
	}
	if err != nil {
		return Config{}, fmt.Errorf("Error Found: %s\nProvide a full set of credentials for a single IAAS.", err)
	}

	if config.StateFileDir == "" {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			return Config{}, err
		}
		config.StateFileDir = dir
	}

	fmt.Printf("Using state-dir: %s\n", config.StateFileDir)

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
	if config.AzureClientID == "" {
		return errors.New("azure client id is missing")
	}

	if config.AzureClientSecret == "" {
		return errors.New("azure client secret is missing")
	}

	if config.AzureRegion == "" {
		return errors.New("azure region is missing")
	}

	if config.AzureSubscriptionID == "" {
		return errors.New("azure subscription id is missing")
	}

	if config.AzureTenantID == "" {
		return errors.New("azure tenant id is missing")
	}

	return nil
}

func validateGCPCreds(config Config) error {
	if config.GCPServiceAccountKey == "" {
		return errors.New("gcp service account key is missing")
	}

	if config.GCPRegion == "" {
		return errors.New("gcp region is missing")
	}

	return nil
}

func validateVSphereCreds(config Config) error {
	if config.VSphereVCenterUser == "" {
		return errors.New("vsphere vcenter user is missing")
	}
	if config.VSphereVCenterPassword == "" {
		return errors.New("vsphere vcenter password is missing")
	}
	if config.VSphereVCenterDC == "" {
		return errors.New("vsphere vcenter datacenter is missing")
	}
	if config.VSphereVCenterDS == "" {
		return errors.New("vsphere vcenter datastore is missing")
	}
	if config.VSphereVCenterCluster == "" {
		return errors.New("vsphere vcenter cluster is missing")
	}
	if config.VSphereVCenterRP == "" {
		return errors.New("vsphere vcenter resource pool is missing")
	}
	if config.VSphereVCenterVMs == "" {
		return errors.New("vsphere vcenter vms is missing")
	}
	if config.VSphereVCenterDisks == "" {
		return errors.New("vsphere vcenter disks is missing")
	}
	if config.VSphereVCenterTemplates == "" {
		return errors.New("vsphere vcenter templates is missing")
	}
	if config.VSphereNetwork == "" {
		return errors.New("vsphere network name is missing")
	}
	return nil
}

func validateOpenStackCreds(config Config) error {
	if config.OpenStackUsername == "" {
		return errors.New("OpenStack user name is missing")
	}
	if config.OpenStackPassword == "" {
		return errors.New("OpenStack password is missing")
	}
	if config.OpenStackAuthURL == "" {
		return errors.New("OpenStack auth URL is missing")
	}
	if config.OpenStackAZ == "" {
		return errors.New("OpenStack AZ is missing")
	}
	if config.OpenStackNetworkID == "" {
		return errors.New("OpenStack network ID is missing")
	}
	if config.OpenStackNetworkName == "" {
		return errors.New("OpenStack network name is missing")
	}
	if config.OpenStackProject == "" {
		return errors.New("OpenStack project is missing")
	}
	if config.OpenStackDomain == "" {
		return errors.New("OpenStack domain is missing")
	}
	if config.OpenStackRegion == "" {
		return errors.New("OpenStack region is missing")
	}
	return nil
}

func loadConfigFromEnvVars() Config {
	return Config{
		IAAS: os.Getenv("BBL_IAAS"),

		StemcellURL: os.Getenv("STEMCELL_URL"),

		AWSAccessKeyID:     os.Getenv("BBL_AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("BBL_AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          os.Getenv("BBL_AWS_REGION"),

		AzureClientID:       os.Getenv("BBL_AZURE_CLIENT_ID"),
		AzureClientSecret:   os.Getenv("BBL_AZURE_CLIENT_SECRET"),
		AzureRegion:         os.Getenv("BBL_AZURE_REGION"),
		AzureSubscriptionID: os.Getenv("BBL_AZURE_SUBSCRIPTION_ID"),
		AzureTenantID:       os.Getenv("BBL_AZURE_TENANT_ID"),

		GCPServiceAccountKey: os.Getenv("BBL_GCP_SERVICE_ACCOUNT_KEY"),
		GCPRegion:            os.Getenv("BBL_GCP_REGION"),

		VSphereNetwork:          os.Getenv("BBL_VSPHERE_NETWORK"),
		VSphereSubnetCIDR:       os.Getenv("BBL_VSPHERE_SUBNET_CIDR"),
		VSphereVCenterIP:        os.Getenv("BBL_VSPHERE_VCENTER_IP"),
		VSphereVCenterUser:      os.Getenv("BBL_VSPHERE_VCENTER_USER"),
		VSphereVCenterPassword:  os.Getenv("BBL_VSPHERE_VCENTER_PASSWORD"),
		VSphereVCenterDC:        os.Getenv("BBL_VSPHERE_VCENTER_DC"),
		VSphereVCenterCluster:   os.Getenv("BBL_VSPHERE_VCENTER_CLUSTER"),
		VSphereVCenterRP:        os.Getenv("BBL_VSPHERE_VCENTER_RP"),
		VSphereVCenterDS:        os.Getenv("BBL_VSPHERE_VCENTER_DS"),
		VSphereVCenterDisks:     os.Getenv("BBL_VSPHERE_VCENTER_DISKS"),
		VSphereVCenterVMs:       os.Getenv("BBL_VSPHERE_VCENTER_VMS"),
		VSphereVCenterTemplates: os.Getenv("BBL_VSPHERE_VCENTER_TEMPLATES"),

		OpenStackAuthURL:     os.Getenv("BBL_OPENSTACK_AUTH_URL"),
		OpenStackAZ:          os.Getenv("BBL_OPENSTACK_AZ"),
		OpenStackNetworkName: os.Getenv("BBL_OPENSTACK_NETWORK_NAME"),
		OpenStackNetworkID:   os.Getenv("BBL_OPENSTACK_NETWORK_ID"),
		OpenStackPassword:    os.Getenv("BBL_OPENSTACK_PASSWORD"),
		OpenStackUsername:    os.Getenv("BBL_OPENSTACK_USERNAME"),
		OpenStackProject:     os.Getenv("BBL_OPENSTACK_PROJECT"),
		OpenStackDomain:      os.Getenv("BBL_OPENSTACK_DOMAIN"),
		OpenStackRegion:      os.Getenv("BBL_OPENSTACK_REGION"),

		StateFileDir:   os.Getenv("BBL_STATE_DIR"),
		BBLStateBucket: os.Getenv("BBL_STATE_BUCKET"),
	}
}

func SkipUnless(match string) {
	test := os.Getenv("RUN_TEST")
	if test != "" && test != match {
		Skip(fmt.Sprintf("RUN_TEST: %s", test))
	}
}
