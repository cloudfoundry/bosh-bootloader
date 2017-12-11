package commands

import "fmt"

const (
	Credentials = `
  --aws-access-key-id        AWS Access Key ID to use       also: $BBL_AWS_ACCESS_KEY_ID
  --aws-secret-access-key    AWS Secret Access Key to use   also: $BBL_AWS_SECRET_ACCESS_KEY
  --aws-region               AWS Region to use              also: $BBL_AWS_REGION

  --gcp-service-account-key  GCP Service Access Key to use  also: $BBL_GCP_SERVICE_ACCOUNT_KEY
  --gcp-region               GCP Region to use              also: $BBL_GCP_REGION

  --azure-subscription-id    Azure Subscription ID to use   also: $BBL_AZURE_SUBSCRIPTION_ID
  --azure-tenant-id          Azure Tenant ID to use         also: $BBL_AZURE_TENANT_ID
  --azure-client-id          Azure Client ID to use         also: $BBL_AZURE_CLIENT_ID
  --azure-client-secret      Azure Client Secret to use     also: $BBL_AZURE_CLIENT_SECRET
  --azure-region             Azure Region to use            also: $BBL_AZURE_REGION`

	requiresCredentials = `

  Credentials for your IaaS are required:`

	LBUsage = `

  Load Balancer options:
  --lb-type                  Load balancer(s) type: "concourse" or "cf"
  [--cert]                   Path to SSL certificate (supported when type="cf")
  [--key]                    Path to SSL certificate key (supported when type="cf")
  [--chain]                  Path to SSL certificate chain (supported when iaas="aws")
  --lb-domain                Creates a DNS zone and records for the given domain (supported when type="cf")`

	PlanCommandUsage = `Populates a state directory with the latest config without applying it

  --iaas                     IAAS to deploy your BOSH director onto: "aws", "azure", "gcp"   also: $BBL_IAAS
  --name                     Name to assign to your BOSH director (optional)                 also: $BBL_ENV_NAME
`

	UpCommandUsage = `Deploys BOSH director on an IAAS

  --iaas                     IAAS to deploy your BOSH director onto: "aws", "azure", "gcp"   also: $BBL_IAAS
  --name                     Name to assign to your BOSH director (optional)                 also: $BBL_ENV_NAME
`

	DestroyCommandUsage = `Tears down BOSH director infrastructure

  [--no-confirm]       Do not ask for confirmation (optional)
  [--skip-if-missing]  Gracefully exit if there is no state file (optional)`

	CreateLBsCommandUsage = `Attaches load balancer(s) with a certificate, key, and optional chain

  --type              Load balancer(s) type: "concourse" or "cf"
  [--cert]            Path to SSL certificate (supported when type="cf")
  [--key]             Path to SSL certificate key (supported when type="cf")
  [--chain]           Path to SSL certificate chain (supported when iaas="aws")
  [--domain]          Creates a DNS zone and records for the given domain (supported when type="cf")`

	DeleteLBsCommandUsage = `Deletes load balancer(s)

  [--skip-if-missing]  Skips deleting load balancer(s) if it is not attached (optional)`

	LBsCommandUsage = "Prints attached load balancer(s)"

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

	BOSHDeploymentVarsCommandUsage = "Prints required variables for BOSH deployment"

	JumpboxDeploymentVarsCommandUsage = "Prints required variables for jumpbox deployment"

	CloudConfigUsage = "Prints suggested cloud configuration for BOSH environment"
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

func (CreateLBs) Usage() string {
	return fmt.Sprintf("%s%s%s", CreateLBsCommandUsage, requiresCredentials, Credentials)
}

func (DeleteLBs) Usage() string {
	return fmt.Sprintf("%s%s%s", DeleteLBsCommandUsage, requiresCredentials, Credentials)
}

func (LBs) Usage() string { return LBsCommandUsage }

func (Version) Usage() string { return VersionCommandUsage }

func (Usage) Usage() string { return UsageCommandUsage }

func (PrintEnv) Usage() string { return PrintEnvCommandUsage }

func (LatestError) Usage() string { return LatestErrorCommandUsage }

func (CloudConfig) Usage() string { return CloudConfigUsage }

func (BOSHDeploymentVars) Usage() string { return BOSHDeploymentVarsCommandUsage }

func (JumpboxDeploymentVars) Usage() string { return JumpboxDeploymentVarsCommandUsage }

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
