package commands

const (
	requiresCredentials = `

  Credentials for your IaaS are required:
  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)`

	certKeyRequirements = `

  --cert/--key are required for cf LBs and are not required or used for concourse LBs.`

	PlanCommandUsage = `Populates a state directory with the latest config without applying it

  --iaas                     IAAS to deploy your BOSH director onto. Valid options: "aws", "azure", "gcp" (Defaults to environment variable BBL_IAAS)
  [--name]                   Name to assign to your BOSH director (optional, will be randomly generated)
  [--ops-file]               Path to BOSH ops file (optional)
  [--no-director]            Skips creating BOSH environment

  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)
  --aws-region               AWS Region to use (Defaults to environment variable BBL_AWS_REGION)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)
  --gcp-region               GCP Region to use (Defaults to environment variable BBL_GCP_REGION)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)
  --azure-location           Azure Location to use (Defaults to environment variable BBL_AZURE_LOCATION)`

	UpCommandUsage = `Deploys BOSH director on an IAAS

  --iaas                     IAAS to deploy your BOSH director onto. Valid options: "aws", "azure", "gcp" (Defaults to environment variable BBL_IAAS)
  [--name]                   Name to assign to your BOSH director (optional, will be randomly generated)
  [--ops-file]               Path to BOSH ops file (optional)
  [--no-director]            Skips creating BOSH environment

  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)
  --aws-region               AWS Region to use (Defaults to environment variable BBL_AWS_REGION)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)
  --gcp-region               GCP Region to use (Defaults to environment variable BBL_GCP_REGION)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)
  --azure-region             Azure Region to use (Defaults to environment variable BBL_AZURE_REGION)`

	DestroyCommandUsage = `Tears down BOSH director infrastructure

  [--no-confirm]       Do not ask for confirmation (optional)
  [--skip-if-missing]  Gracefully exit if there is no state file (optional)` + requiresCredentials

	CreateLBsCommandUsage = `Attaches load balancer(s) with a certificate, key, and optional chain

  --type              Load balancer(s) type. Valid options: "concourse" or "cf"
  [--cert]            Path to SSL certificate (conditionally required; refer to table below)
  [--key]             Path to SSL certificate key (conditionally required; refer to table below)
  [--chain]           Path to SSL certificate chain (optional; only supported on aws)
  [--domain]          Creates a DNS zone and records for the given domain (supported when type="cf")` + requiresCredentials + certKeyRequirements

	DeleteLBsCommandUsage = `Deletes load balancer(s)

  [--skip-if-missing]  Skips deleting load balancer(s) if it is not attached (optional)` + requiresCredentials

	LBsCommandUsage = "Prints attached load balancer(s)"

	VersionCommandUsage = "Prints version"

	UsageCommandUsage = "Prints helpful message for the given command"

	EnvIdCommandUsage = "Prints environment ID"

	SSHKeyCommandUsage = "Prints SSH private key for the jumpbox."

	DirectorSSHKeyCommandUsage = "Prints SSH private key for the director."

	RotateCommandUsage = "Rotates SSH key for the jumpbox user." + requiresCredentials

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

func (Up) Usage() string { return UpCommandUsage }

func (Plan) Usage() string { return PlanCommandUsage }

func (Destroy) Usage() string { return DestroyCommandUsage }

func (CreateLBs) Usage() string { return CreateLBsCommandUsage }

func (DeleteLBs) Usage() string { return DeleteLBsCommandUsage }

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

func (Rotate) Usage() string { return RotateCommandUsage }

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
