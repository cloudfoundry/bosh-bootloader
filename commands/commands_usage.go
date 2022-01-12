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
  --vsphere-subnet-cidr              vSphere Subnet CIDR              env: $BBL_VSPHERE_SUBNET_CIDR
  --vsphere-vcenter-disks            vSphere vCenter Disks            env: $BBL_VSPHERE_VCENTER_DISKS
  --vsphere-vcenter-templates        vSphere vCenter Templates        env: $BBL_VSPHERE_VCENTER_TEMPLATES
  --vsphere-vcenter-vms              vSphere vCenter VMs              env: $BBL_VSPHERE_VCENTER_VMS

  --openstack-auth-url               OpenStack Auth URL               env: $BBL_OPENSTACK_AUTH_URL
  --openstack-az                     OpenStack Availability Zone      env: $BBL_OPENSTACK_AZ
  --openstack-network-id             OpenStack Network ID             env: $BBL_OPENSTACK_NETWORK_ID
  --openstack-network-name           OpenStack Network Name           env: $BBL_OPENSTACK_NETWORK_NAME
  --openstack-password               OpenStack Password               env: $BBL_OPENSTACK_PASSWORD
  --openstack-username               OpenStack Username               env: $BBL_OPENSTACK_USERNAME
  --openstack-project                OpenStack Project                env: $BBL_OPENSTACK_PROJECT
  --openstack-domain                 OpenStack Domain                 env: $BBL_OPENSTACK_DOMAIN
  --openstack-region                 OpenStack Region                 env: $BBL_OPENSTACK_REGION
  --openstack-cacert-file            OpenStack CA Cert File           env: $BBL_OPENSTACK_CACERT_FILE
  --openstack-insecure               OpenStack Insecure               env: $BBL_OPENSTACK_INSECURE
  --openstack-dns-name-server        OpenStack DNS Name Servers       env: $BBL_OPENSTACK_DNS_NAME_SERVERS

  --cloudstack-endpoint              CloudStack Endpoint              env: $BBL_CLOUDSTACK_ENDPOINT
  --cloudstack-secret-access-key     CloudStack Secret Access Key     env: $BBL_CLOUDSTACK_SECRET_ACCESS_KEY
  --cloudstack-api-key               CloudStack Api Key               env: $BBL_CLOUDSTACK_API_KEY
  --cloudstack-zone                  CloudStack Zone                  env: $BBL_CLOUDSTACK_ZONE
  --cloudstack-secure                CloudStack Activate sec group    env: $BBL_CLOUDSTACK_SECURE
  --cloudstack-iso-segment           CloudStack Activate iso segemnt  env: $BBL_CLOUDSTACK_ISO_SEGMENT`

	requiresCredentials = `

  Credentials for your IaaS are required:`

	LBUsage = `

  Load Balancer options:
  --lb-type                  Load balancer(s) type: "concourse" or "cf"
  --lb-cert                  Path to SSL certificate (supported when type="cf")
  --lb-key                   Path to SSL certificate key (supported when type="cf")
  --lb-domain                Creates a DNS zone and records for the given domain (supported when type="cf")`

	PlanCommandUsage = `Populates a state directory with the latest config without applying it

  --iaas                     IAAS to deploy your BOSH director onto: "aws", "azure", "gcp", "vsphere", "cloudstack"   env: $BBL_IAAS
  --name                     Name to assign to your BOSH director (optional)                            env: $BBL_ENV_NAME
`

	UpCommandUsage = `Deploys BOSH director on an IAAS

  --iaas                     IAAS to deploy your BOSH director onto: "aws", "azure", "gcp", "vsphere"   env: $BBL_IAAS
  --name                     Name to assign to your BOSH director (optional)                            env: $BBL_ENV_NAME
`

	DestroyCommandUsage = `Tears down BOSH director infrastructure

  [--no-confirm]       Do not ask for confirmation (optional)`

	CleanupLeftoversCommandUsage = `Cleans up orphaned IAAS resources

  --filter            Only delete resources with this string in their name
  --dry-run           List all resources without deleting any`

	LBsCommandUsage = "Prints attached load balancer(s)"

	OutputsCommandUsage = "Prints the outputs from terraform."

	VersionCommandUsage = "Prints version"

	UsageCommandUsage = "Prints helpful message for the given command"

	EnvIdCommandUsage = "Prints environment ID"

	SSHKeyCommandUsage = "Prints SSH private key for the jumpbox."

	DirectorSSHKeyCommandUsage = "Prints SSH private key for the director."

	SSHCommandUsage = `Opens an SSH connection to the director or the jumpbox.

  --jumpbox                Open a connection to the jumpbox
  --director               Open a connection to the director
  --cmd                    Execute a command on the director (jumpbox not supported)
`

	RotateCommandUsage = "Rotates SSH key for the jumpbox user."

	JumpboxAddressCommandUsage = "Prints BOSH jumpbox address"

	DirectorUsernameCommandUsage = "Prints BOSH director username"

	DirectorPasswordCommandUsage = "Prints BOSH director password"

	DirectorAddressCommandUsage = "Prints BOSH director address"

	DirectorCACertCommandUsage = "Prints BOSH director CA certificate"

	PrintEnvCommandUsage = `Prints required BOSH environment variables.

  --shell-type             Prints for the given shell (posix|powershell|yaml)
  --metadata-file          Read from Toolsmiths metadata file instead of bbl state
`
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

func (Validate) Usage() string { return "" }

func (s SSHKey) Usage() string {
	if s.Director {
		return DirectorSSHKeyCommandUsage
	}
	return SSHKeyCommandUsage
}

func (s SSH) Usage() string {
	return SSHCommandUsage
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
