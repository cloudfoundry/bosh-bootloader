package commands

const (
	UpCommandUsage = `Deploys BOSH director on an IAAS

  --iaas                     IAAS to deploy your BOSH Director onto. Valid options: "GCP", "AWS"
  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)
  --aws-region               AWS region to use (Defaults to environment variable BBL_AWS_REGION)`

	DestroyCommandUsage = `Tears down BOSH director infrastructure

  [--no-confirm]       Do not ask for confirmation (optional)
  [--skip-if-missing]  Gracefully exit if there is no state file (optional)`

	CreateLBsCommandUsage = `Attaches load balancer(s) with a certificate, key, and optional chain

  --type              Load balancer(s) type. Valid options: "concourse" or "cf"
  --cert              Path to SSL certificate
  --key               Path to SSL certificate key
  [--chain]           Path to SSL certificate chain (optional)
  [--skip-if-exists]  Skips creating load balancer(s) if it is already attached (optional)`

	UpdateLBsCommandUsage = `Updates load balancer(s) with the supplied certificate, key, and optional chain

  --cert               Path to SSL certificate
  --key                Path to SSL certificate key
  [--chain]            Path to SSL certificate chain (optional)
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
)

func (Up) Usage() string { return UpCommandUsage }

func (Destroy) Usage() string { return DestroyCommandUsage }

func (CreateLBs) Usage() string { return CreateLBsCommandUsage }

func (UpdateLBs) Usage() string { return UpdateLBsCommandUsage }

func (DeleteLBs) Usage() string { return DeleteLBsCommandUsage }

func (LBs) Usage() string { return LBsCommandUsage }

func (Version) Usage() string { return VersionCommandUsage }

func (Usage) Usage() string { return UsageCommandUsage }

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
