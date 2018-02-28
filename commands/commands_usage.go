package commands

import "fmt"

const (
	Credentials = `
  --aws-access-key-id                AWS Access Key ID                env: $BBL_AWS_ACCESS_KEY_ID
  --aws-secret-access-key            AWS Secret Access Key            env: $BBL_AWS_SECRET_ACCESS_KEY
  --aws-region                       AWS Region                       env: $BBL_AWS_REGION

  --gcp-service-account-key          GCP Service Access Key to use    env: $BBL_GCP_SERVICE_ACCOUNT_KEY
  --gcp-region                       GCP Region to use                env: $BBL_GCP_REGION

  --azure-subscription-id            Azure Subscription ID            env: $BBL_AZURE_SUBSCRIPTION_ID
  --azure-tenant-id                  Azure Tenant ID                  env: $BBL_AZURE_TENANT_ID
  --azure-client-id                  Azure Client ID                  env: $BBL_AZURE_CLIENT_ID
  --azure-client-secret              Azure Client Secret              env: $BBL_AZURE_CLIENT_SECRET
  --azure-region                     Azure Region                     env: $BBL_AZURE_REGION

  --vsphere-vcenter-user             vSphere vCenter User             env: $BBL_VSPHERE_VCENTER_USER
  --vsphere-vcenter-password         vSphere vCenter Password         env: $BBL_VSPHERE_VCENTER_PASSWORD
  --vsphere-vcenter-ip               vSphere vCenter IP               env: $BBL_VSPHERE_VCENTER_IP
  --vsphere-vcenter-dc               vSphere vCenter Datacenter       env: $BBL_VSPHERE_VCENTER_DC
  --vsphere-vcenter-cluster          vSphere vCenter Cluster          env: $BBL_VSPHERE_VCENTER_CLUSTER
  --vsphere-vcenter-rp               vSphere vCenter Resource Pool    env: $BBL_VSPHERE_VCENTER_RP
  --vsphere-network                  vSphere Network                  env: $BBL_VSPHERE_NETWORK
  --vsphere-vcenter-ds               vSphere vCenter Datastore        env: $BBL_VSPHERE_VCENTER_DS
  --vsphere-subnet                   vSphere Subnet                   env: $BBL_VSPHERE_SUBNET

  --openstack-internal-cidr          OpenStack Internal CIDR          env: $BBL_OPENSTACK_INTERNAL_CIDR
  --openstack-external-ip            OpenStack External IP            env: $BBL_OPENSTACK_EXTERNAL_IP
  --openstack-auth-url               OpenStack Auth URL               env: $BBL_OPENSTACK_AUTH_URL
  --openstack-az                     OpenStack Availability Zone      env: $BBL_OPENSTACK_AZ
  --openstack-default-key-name       OpenStack Default Key Name       env: $BBL_OPENSTACK_DEFAULT_KEY_NAME
  --openstack-default-security-group OpenStack Default Security Group env: $BBL_OPENSTACK_DEFAULT_SECURITY_GROUP
  --openstack-network-id             OpenStack Network ID             env: $BBL_OPENSTACK_NETWORK_ID
  --openstack-password               OpenStack Password               env: $BBL_OPENSTACK_PASSWORD
  --openstack-username               OpenStack Username               env: $BBL_OPENSTACK_USERNAME
  --openstack-project                OpenStack Project                env: $BBL_OPENSTACK_PROJECT
  --openstack-domain                 OpenStack Domain                 env: $BBL_OPENSTACK_DOMAIN
  --openstack-region                 OpenStack Region                 env: $BBL_OPENSTACK_REGION
  --openstack-private-key            OpenStack Private Key            env: $BBL_OPENSTACK_PRIVATE_KEY`

	requiresCredentials = `

  Credentials for your IaaS are required:`

	LBUsage = `

  Load Balancer options:
  --lb-type                  Load balancer(s) type: "concourse" or "cf"
  --lb-cert                  Path to SSL certificate (supported when type="cf")
  --lb-key                   Path to SSL certificate key (supported when type="cf")
  --lb-chain                 Path to SSL certificate chain (supported when iaas="aws")
  --lb-domain                Creates a DNS zone and records for the given domain (supported when type="cf")`

	PlanCommandUsage = `Populates a state directory with the latest config without applying it

  --iaas                     IAAS to deploy your BOSH director onto: "aws", "azure", "gcp", "vsphere"   env: $BBL_IAAS
  --name                     Name to assign to your BOSH director (optional)                            env: $BBL_ENV_NAME
`

	UpCommandUsage = `Deploys BOSH director on an IAAS

  --iaas                     IAAS to deploy your BOSH director onto: "aws", "azure", "gcp", "vsphere"   env: $BBL_IAAS
  --name                     Name to assign to your BOSH director (optional)                            env: $BBL_ENV_NAME
`

	DestroyCommandUsage = `Tears down BOSH director infrastructure

  [--no-confirm]       Do not ask for confirmation (optional)
  [--skip-if-missing]  Gracefully exit if there is no state file (optional)`

	CleanupLeftoversCommandUsage = `Cleans up orphaned IAAS resources

  --filter            Only delete resources with this string in their name`

	LBsCommandUsage = "Prints attached load balancer(s)"

	OutputsCommandUsage = "Prints the outputs from terraform."

	VersionCommandUsage = "Prints version"

	UsageCommandUsage = "Prints helpful message for the given command"

	EnvIdCommandUsage = "Prints environment ID"

	SSHKeyCommandUsage = "Prints SSH private key for the jumpbox."

	DirectorSSHKeyCommandUsage = "Prints SSH private key for the director."

	RotateCommandUsage = "Rotates SSH key for the jumpbox user."

	JumpboxAddressCommandUsage = "Prints BOSH jumpbox address"

	DirectorUsernameCommandUsage = "Prints BOSH director username"

	DirectorPasswordCommandUsage = "Prints BOSH director password"

	DirectorAddressCommandUsage = "Prints BOSH director address"

	DirectorCACertCommandUsage = "Prints BOSH director CA certificate"

	PrintEnvCommandUsage = "Prints required BOSH environment variables"

	LatestErrorCommandUsage = "Prints the output from the latest call to terraform"
)

func (Up) Usage() string {
	return fmt.Sprintf("%s%s%s", UpCommandUsage, Credentials, LBUsage)
}

func (Plan) Usage() string {
	return fmt.Sprintf("%s%s%s", PlanCommandUsage, Credentials, LBUsage)
}

func (Destroy) Usage() string {
	return fmt.Sprintf("%s%s%s", DestroyCommandUsage, requiresCredentials, Credentials)
}

func (Rotate) Usage() string {
	return fmt.Sprintf("%s%s%s", RotateCommandUsage, requiresCredentials, Credentials)
}

func (LBs) Usage() string { return LBsCommandUsage }

func (Outputs) Usage() string { return OutputsCommandUsage }

func (Version) Usage() string { return VersionCommandUsage }

func (Usage) Usage() string { return UsageCommandUsage }

func (PrintEnv) Usage() string { return PrintEnvCommandUsage }

func (LatestError) Usage() string { return LatestErrorCommandUsage }

func (s SSHKey) Usage() string {
	if s.Director {
		return DirectorSSHKeyCommandUsage
	}
	return SSHKeyCommandUsage
}

func (s StateQuery) Usage() string {
	switch s.propertyName {
	case EnvIDPropertyName:
		return EnvIdCommandUsage
	case JumpboxAddressPropertyName:
		return JumpboxAddressCommandUsage
	case DirectorUsernamePropertyName:
		return DirectorUsernameCommandUsage
	case DirectorPasswordPropertyName:
		return DirectorPasswordCommandUsage
	case DirectorAddressPropertyName:
		return DirectorAddressCommandUsage
	case DirectorCACertPropertyName:
		return DirectorCACertCommandUsage
	}
	return ""
}
