package commands

const (
	UpCommandUsage = `Deploys a BOSH Director on AWS

--aws-access-key-id        AWS Access Key ID to use
--aws-secret-access-key    AWS Secret Access Key to use
--aws-region               AWS region to use`

	DestroyCommandUsage = `Tears down a BOSH Director environment on AWS

--no-confirm  Do not ask for confirmation (optional)`

	CreateLBsCommandUsage = `Attaches a load balancer with a certificate, key, and optional chain

--type            Load balancer type. Valid options: "concourse" or "cf"
--cert            Path to SSL certificate
--key             Path to SSL certificate key
--chain           Path to SSL certificate chain (optional)
--skip-if-exists  Skips creating load balancer if it is already attached (optional)`

	UpdateLBsCommandUsage = `Updates a load balancer with the supplied certificate, key, and optional chain

--cert             Path to SSL certificate
--key              Path to SSL certificate key
--chain            Path to SSL certificate chain (optional)
--skip-if-missing  Skips updating load balancer if it is not attached (optional)`

	DeleteLBsCommandUsage = `Deletes the load balancers

--skip-if-missing  Skips deleting load balancer if it is not attached (optional)`

	LBsCommandUsage = "Lists attached load balancers"

	VersionCommandUsage = "Prints version"

	UsageCommandUsage = "Prints helpful message for the given command"

	EnvIdCommandUsage = "Prints the environment ID"

	SSHKeyCommandUsage = "Prints the SSH private key"

	DirectorUsernameCommandUsage = "Prints the BOSH director username"

	DirectorPasswordCommandUsage = "Prints the BOSH director password"

	DirectorAddressCommandUsage = "Prints the BOSH director address"

	BOSHCACertCommandUsage = "Prints the BOSH director CA certificate"
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
	case BOSHCACertPropertyName:
		return BOSHCACertCommandUsage
	}
	return ""
}
