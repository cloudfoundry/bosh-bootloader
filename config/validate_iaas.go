package config

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

func ValidateIAAS(state storage.State) error {
	var err error
	switch state.IAAS {
	case "aws":
		err = aws(state.AWS)
	case "azure":
		err = azure(state.Azure)
	case "gcp":
		err = gcp(state.GCP)
	case "vsphere":
		err = vsphere(state.VSphere)
	case "openstack":
		err = openstack(state.OpenStack)
	case "cloudstack":
		err = cloudstack(state.CloudStack)
	default:
		err = errors.New("--iaas [gcp, aws, azure, vsphere, openstack, cloudstack] must be provided or BBL_IAAS must be set")
	}

	if err != nil {
		return fmt.Errorf("\n\n%s\n", err)
	}

	return nil
}

const CRED_ERROR = "Missing %s. To see all required credentials run `bbl plan --help`."

func aws(state storage.AWS) error {
	if state.AccessKeyID == "" {
		return fmt.Errorf(CRED_ERROR, "--aws-access-key-id")
	}
	if state.SecretAccessKey == "" {
		return fmt.Errorf(CRED_ERROR, "--aws-secret-access-key")
	}
	if state.Region == "" {
		return fmt.Errorf(CRED_ERROR, "--aws-region")
	}
	return nil
}

func azure(state storage.Azure) error {
	if state.ClientID == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-client-id")
	}
	if state.ClientSecret == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-client-secret")
	}
	if state.Region == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-region")
	}
	if state.SubscriptionID == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-subscription-id")
	}
	if state.TenantID == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-tenant-id")
	}
	return nil
}

func gcp(state storage.GCP) error {
	if state.ServiceAccountKey == "" {
		return fmt.Errorf(CRED_ERROR, "--gcp-service-account-key")
	}
	if state.Region == "" {
		return fmt.Errorf(CRED_ERROR, "--gcp-region")
	}
	return nil
}

func openstack(state storage.OpenStack) error {
	if state.AuthURL == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-auth-url")
	}
	if state.AZ == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-az")
	}
	if state.NetworkID == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-network-id")
	}
	if state.NetworkName == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-network-name")
	}
	if state.Username == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-username")
	}
	if state.Password == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-password")
	}
	if state.Project == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-project")
	}
	if state.Domain == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-domain")
	}
	if state.Region == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-region")
	}
	return nil
}

func vsphere(state storage.VSphere) error {
	if state.VCenterUser == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-user")
	}
	if state.VCenterPassword == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-password")
	}
	if state.VCenterIP == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-ip")
	}
	if state.VCenterDC == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-dc")
	}
	if state.VCenterRP == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-rp")
	}
	if state.VCenterDS == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-ds")
	}
	if state.VCenterCluster == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-cluster")
	}
	if state.Network == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-network")
	}
	if state.SubnetCIDR == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-subnet-cidr")
	}
	return nil
}

func cloudstack(state storage.CloudStack) error {
	if state.Endpoint == "" {
		return fmt.Errorf(CRED_ERROR, "--cloudstack-endpoint")
	}
	if state.ApiKey == "" {
		return fmt.Errorf(CRED_ERROR, "--cloudstack-api-key")
	}
	if state.SecretAccessKey == "" {
		return fmt.Errorf(CRED_ERROR, "--cloudstack-secret-access-key")
	}
	if state.Zone == "" {
		return fmt.Errorf(CRED_ERROR, "--cloudstack-zone")
	}
	return nil
}
