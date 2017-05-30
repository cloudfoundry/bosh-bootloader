package commands

const (
	UpCommandUsage = `Deploys BOSH director on an IAAS

  --iaas                     IAAS to deploy your BOSH Director onto. Valid options: "gcp", "aws" (Defaults to environment variable BBL_IAAS)
  [--name]                   Name to assign to your BOSH Director (optional, will be randomly generated)
  [--ops-file]               Path to BOSH ops file (optional)
  [--jumpbox]                Deploy your BOSH Director behind a jumpbox (supported when iaas="gcp")
  [--no-director]            Skips creating BOSH environment

  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)
  --aws-region               AWS region to use (Defaults to environment variable BBL_AWS_REGION)
  [--aws-bosh-az]            AWS availability zone to use for BOSH director (Defaults to environment variable BBL_AWS_BOSH_AZ)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)
  --gcp-project-id           GCP Project ID to use (Defaults to environment variable BBL_GCP_PROJECT_ID)
  --gcp-zone                 GCP Zone to use (Defaults to environment variable BBL_GCP_ZONE)
  --gcp-region               GCP Region to use (Defaults to environment variable BBL_GCP_REGION)`

	DestroyCommandUsage = `Tears down BOSH director infrastructure

  [--no-confirm]       Do not ask for confirmation (optional)
  [--skip-if-missing]  Gracefully exit if there is no state file (optional)`

	CreateLBsCommandUsage = `Attaches load balancer(s) with a certificate, key, and optional chain

  --type              Load balancer(s) type. Valid options: "concourse" or "cf"
  [--cert]            Path to SSL certificate (conditionally required; refer to table below)
  [--key]             Path to SSL certificate key (conditionally required; refer to table below)
  [--chain]           Path to SSL certificate chain (optional; applicable if --cert/--key are required; refer to table below)
  [--domain]          Creates a nameserver with a zone for given domain (supported when type="cf")
  [--skip-if-exists]  Skips creating load balancer(s) if it is already attached (optional)

  --cert/--key requirements:
  ------------------------------
  |     | cf       | concourse |
  ------------------------------
  | aws | required | required  |
  ------------------------------
  | gcp | required | n/a       |
  ------------------------------`

	UpdateLBsCommandUsage = `Updates load balancer(s) with the supplied certificate, key, and optional chain

  --cert               Path to SSL certificate
  --key                Path to SSL certificate key
  [--chain]            Path to SSL certificate chain (optional)
  [--domain]           Updates domain in the nameserver zone (supported when type="cf", optional)
  [--skip-if-missing]  Skips updating load balancer(s) if it is not attached (optional)`

	DeleteLBsCommandUsage = `Deletes load balancer(s)

  [--skip-if-missing]  Skips deleting load balancer(s) if it is not attached (optional)`

	LBsCommandUsage = "Prints attached load balancer(s)"

	VersionCommandUsage = "Prints version"

	UsageCommandUsage = "Prints helpful message for the given command"

	EnvIdCommandUsage = "Prints environment ID"

	SSHKeyCommandUsage = "Prints SSH private key for the jumpbox user. This can be used to ssh to the director/use the director as a gateway host."

	RotateCommandUsage = "Rotates the keypair for BOSH"

	DirectorUsernameCommandUsage = "Prints BOSH director username"

	DirectorPasswordCommandUsage = "Prints BOSH director password"

	DirectorAddressCommandUsage = "Prints BOSH director address"

	DirectorCACertCommandUsage = "Prints BOSH director CA certificate"

	PrintEnvCommandUsage = "Prints required BOSH environment variables"

	LatestErrorCommandUsage = "Prints the output from the latest call to terraform"

	BOSHDeploymentVarsCommandUsage = "Prints required variables for BOSH deployment"

	CloudConfigUsage = "Prints suggested cloud configuration for BOSH environment"
)

func (Up) Usage() string { return UpCommandUsage }

func (Destroy) Usage() string { return DestroyCommandUsage }

func (CreateLBs) Usage() string { return CreateLBsCommandUsage }

func (UpdateLBs) Usage() string { return UpdateLBsCommandUsage }

func (DeleteLBs) Usage() string { return DeleteLBsCommandUsage }

func (LBs) Usage() string { return LBsCommandUsage }

func (Version) Usage() string { return VersionCommandUsage }

func (Usage) Usage() string { return UsageCommandUsage }

func (PrintEnv) Usage() string { return PrintEnvCommandUsage }

func (LatestError) Usage() string { return LatestErrorCommandUsage }

func (CloudConfig) Usage() string { return CloudConfigUsage }

func (BOSHDeploymentVars) Usage() string { return BOSHDeploymentVarsCommandUsage }

func (Rotate) Usage() string { return RotateCommandUsage }

func (SSHKey) Usage() string { return SSHKeyCommandUsage }

func (s StateQuery) Usage() string {
	switch s.propertyName {
	case EnvIDPropertyName:
		return EnvIdCommandUsage
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
