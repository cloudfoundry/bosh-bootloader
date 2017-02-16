package commands

const (
	UpCommandUsage = `Deploys BOSH director on an IAAS

  --iaas                     IAAS to deploy your BOSH Director onto. Valid options: "gcp", "aws" (Defaults to environment variable BBL_IAAS)
  [--name]                   Name to assign to your BOSH Director (optional, will be randomly generated)
  [--ops-file]               Path to BOSH ops file (optional)

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
  [--cert]            Path to SSL certificate (required when type="cf")
  [--key]             Path to SSL certificate key (required when type="cf")
  [--chain]           Path to SSL certificate chain (optional)
  [--domain]          Creates a nameserver with a zone for given domain
  [--skip-if-exists]  Skips creating load balancer(s) if it is already attached (optional)`

	UpdateLBsCommandUsage = `Updates load balancer(s) with the supplied certificate, key, and optional chain

  --cert               Path to SSL certificate
  --key                Path to SSL certificate key
  [--chain]            Path to SSL certificate chain (optional)
  [--domain]           Updates domain in the nameserver zone (optional)
  [--skip-if-missing]  Skips updating load balancer(s) if it is not attached (optional)`

	DeleteLBsCommandUsage = `Deletes load balancer(s)

  [--skip-if-missing]  Skips deleting load balancer(s) if it is not attached (optional)`

	LBsCommandUsage = "Prints attached load balancer(s)"

	VersionCommandUsage = "Prints version"

	UsageCommandUsage = "Prints helpful message for the given command"

	EnvIdCommandUsage = "Prints environment ID"

	SSHKeyCommandUsage = "Prints SSH private key"

	DirectorUsernameCommandUsage = "Prints BOSH director username"

	DirectorPasswordCommandUsage = "Prints BOSH director password"

	DirectorAddressCommandUsage = "Prints BOSH director address"

	DirectorCACertCommandUsage = "Prints BOSH director CA certificate"

	PrintEnvCommandUsage = "Prints required BOSH environment variables"
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

func (s StateQuery) Usage() string {
	switch s.propertyName {
	case EnvIDPropertyName:
		return EnvIdCommandUsage
	case SSHKeyPropertyName:
		return SSHKeyCommandUsage
	case DirectorUsernamePropertyName:
		return DirectorUsernameCommandUsage
	case DirectorPasswordPropertyName:
		return DirectorPasswordCommandUsage
	case DirectorAddressPropertyName:
		return DirectorAddressCommandUsage
	case DirectorCACertPropertyName:
		return DirectorCACertCommandUsage
	case BOSHCACertPropertyName:
		return DirectorCACertCommandUsage
	}
	return ""
}
