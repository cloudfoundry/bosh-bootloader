# BOSH Bootloader

```
Usage:
  bbl [GLOBAL OPTIONS] COMMAND [OPTIONS]

Global Options:
  --help       [-h]        Prints usage. Use "bbl [command] --help" for more information about a command
  --state-dir  [-s]        Directory containing the bbl state                                            env:"BBL_STATE_DIRECTORY"
  --debug      [-d]        Prints debugging output                                                       env:"BBL_DEBUG"
  --version    [-v]        Prints version
  --no-confirm [-n]        No confirm

Basic Commands: A good place to start
  up                      Deploys BOSH director on an IAAS, creates CF/Concourse load balancers. Updates existing director.
  print-env               All environment variables needed for targeting BOSH. Use with: eval "$(bbl print-env)"

Maintenance Lifecycle Commands:
  destroy                 Tears down BOSH director infrastructure. Cleans up state directory
  rotate                  Rotates SSH key for the jumpbox user
  plan                    Populates a state directory with the latest config without applying it
  cleanup-leftovers       Cleans up orphaned IAAS resources

Environmental Detail Commands: Useful for automation and gaining access
  jumpbox-address         Prints BOSH jumpbox address
  director-address        Prints BOSH director address
  director-username       Prints BOSH director username
  director-password       Prints BOSH director password
  director-ca-cert        Prints BOSH director CA certificate
  env-id                  Prints environment ID
  ssh-key                 Prints jumpbox SSH private key
  director-ssh-key        Prints director SSH private key
  lbs                     Prints load balancer(s) and DNS records
  outputs                 Prints the outputs from terraform

Troubleshooting Commands:
  help                    Prints usage
  version                 Prints version
  latest-error            Prints the output from the latest call to terraform
```
