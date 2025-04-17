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
		return fmt.Errorf("\n\n%s\n", err) //nolint:staticcheck
	}

	return nil
}

const CRED_ERROR = "Missing %s. To see all required credentials run `bbl plan --help`."

func aws(state storage.AWS) error {
	if state.AccessKeyID == "" {
		return fmt.Errorf(CRED_ERROR, "--aws-access-key-id") //nolint:staticcheck
	}
	if state.SecretAccessKey == "" {
		return fmt.Errorf(CRED_ERROR, "--aws-secret-access-key") //nolint:staticcheck
	}
	if state.Region == "" {
		return fmt.Errorf(CRED_ERROR, "--aws-region") //nolint:staticcheck
	}
	return nil
}

func azure(state storage.Azure) error {
	if state.ClientID == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-client-id") //nolint:staticcheck
	}
	if state.ClientSecret == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-client-secret") //nolint:staticcheck
	}
	if state.Region == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-region") //nolint:staticcheck
	}
	if state.SubscriptionID == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-subscription-id") //nolint:staticcheck
	}
	if state.TenantID == "" {
		return fmt.Errorf(CRED_ERROR, "--azure-tenant-id") //nolint:staticcheck
	}
	return nil
}

func gcp(state storage.GCP) error {
	if state.ServiceAccountKey == "" {
		return fmt.Errorf(CRED_ERROR, "--gcp-service-account-key") //nolint:staticcheck
	}
	if state.Region == "" {
		return fmt.Errorf(CRED_ERROR, "--gcp-region") //nolint:staticcheck
	}
	return nil
}

func openstack(state storage.OpenStack) error {
	if state.AuthURL == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-auth-url") //nolint:staticcheck
	}
	if state.AZ == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-az") //nolint:staticcheck
	}
	if state.NetworkID == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-network-id") //nolint:staticcheck
	}
	if state.NetworkName == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-network-name") //nolint:staticcheck
	}
	if state.Username == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-username") //nolint:staticcheck
	}
	if state.Password == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-password") //nolint:staticcheck
	}
	if state.Project == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-project") //nolint:staticcheck
	}
	if state.Domain == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-domain") //nolint:staticcheck
	}
	if state.Region == "" {
		return fmt.Errorf(CRED_ERROR, "--openstack-region") //nolint:staticcheck
	}
	return nil
}

func vsphere(state storage.VSphere) error {
	if state.VCenterUser == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-user") //nolint:staticcheck
	}
	if state.VCenterPassword == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-password") //nolint:staticcheck
	}
	if state.VCenterIP == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-ip") //nolint:staticcheck
	}
	if state.VCenterDC == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-dc") //nolint:staticcheck
	}
	if state.VCenterRP == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-rp") //nolint:staticcheck
	}
	if state.VCenterDS == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-ds") //nolint:staticcheck
	}
	if state.VCenterCluster == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-vcenter-cluster") //nolint:staticcheck
	}
	if state.Network == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-network") //nolint:staticcheck
	}
	if state.SubnetCIDR == "" {
		return fmt.Errorf(CRED_ERROR, "--vsphere-subnet-cidr") //nolint:staticcheck
	}
	return nil
}

func cloudstack(state storage.CloudStack) error {
	if state.Endpoint == "" {
		return fmt.Errorf(CRED_ERROR, "--cloudstack-endpoint") //nolint:staticcheck
	}
	if state.ApiKey == "" {
		return fmt.Errorf(CRED_ERROR, "--cloudstack-api-key") //nolint:staticcheck
	}
	if state.SecretAccessKey == "" {
		return fmt.Errorf(CRED_ERROR, "--cloudstack-secret-access-key") //nolint:staticcheck
	}
	if state.Zone == "" {
		return fmt.Errorf(CRED_ERROR, "--cloudstack-zone") //nolint:staticcheck
	}
	return nil
}
