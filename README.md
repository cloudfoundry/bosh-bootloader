# bosh-bootloader
---

This is a command line utility for standing up a CloudFoundry or Concourse installation
on an IAAS. This CLI supports bootstrapping a CloudFoundry or Concourse installation on AWS and GCP.

* [CI](https://wings.concourse.ci/teams/cf-infrastructure/pipelines/bosh-bootloader)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## Guides

- [Getting Started on AWS](docs/getting-started-aws.md)
- [Deploying Concourse on GCP](docs/concourse.md)
- [Deploying Cloud Foundry on GCP](docs/cloudfoundry.md)
- [Advanced BOSH Configuration](docs/advanced.md)

## Prerequisites

### Install bosh-bootloader using a package manager

**Mac OS X** (using [Homebrew](http://brew.sh/) via the [cloudfoundry tap](https://github.com/cloudfoundry/homebrew-tap)):

```sh
$ brew install cloudfoundry/tap/bbl
```

### Install Dependencies

The following should be installed on your local machine
- BOSH v2 CLI  [BOSH v2 CLI](https://bosh.io/docs/cli-v2.html). Ensure that the binary gets installed as `bosh`; this requires setting the option `--without-bosh2` when installing via Homebrew.
```sh
$ brew install cloudfoundry/tap/bosh-cli --without-bosh2
```
- terraform >= 0.9.7 ([download here](https://www.terraform.io/downloads.html))
- ruby

### Install bosh-bootloader

bosh-bootloader can be installed by downloading the [latest Github release](https://github.com/cloudfoundry/bosh-bootloader/releases/latest).

### Configure AWS

The AWS IAM user that is provided to bbl will need the following policy:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
		"ec2:*",
                "cloudformation:*",
                "elasticloadbalancing:*",
                "route53:*",
                "iam:*",
                "logs:*"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

### Configure GCP

To allow bbl to set up infrastructure a service account must be provided with the
role 'roles/editor'

Example:
```
gcloud iam service-accounts create <service account name>

gcloud iam service-accounts keys create --iam-account='<service account name>@<project id>.iam.gserviceaccount.com' <service account name>.key.json

gcloud projects add-iam-policy-binding <project id> --member='serviceAccount:<service account name>@<project id>.iam.gserviceaccount.com' --role='roles/editor'
```

## Usage

The `bbl` command can be invoked on the command line and will display its usage.

```
$ bbl
Usage:
  bbl [GLOBAL OPTIONS] COMMAND [OPTIONS]

Global Options:
  --help      [-h]       Prints usage
  --state-dir            Directory containing bbl-state.json
  --debug                Prints debugging output
  --version              Prints version

Commands:
  bosh-deployment-vars   Prints required variables for BOSH deployment
  cloud-config           Prints suggested cloud configuration for BOSH environment
  create-lbs             Attaches load balancer(s)
  delete-lbs             Deletes attached load balancer(s)
  destroy                Tears down BOSH director infrastructure
  director-address       Prints BOSH director address
  director-username      Prints BOSH director username
  director-password      Prints BOSH director password
  director-ca-cert       Prints BOSH director CA certificate
  env-id                 Prints environment ID
  latest-error           Prints the output from the latest call to terraform
  print-env              Prints BOSH friendly environment variables
  help                   Prints usage
  lbs                    Prints attached load balancer(s)
  ssh-key                Prints SSH private key
  up                     Deploys BOSH director on an IAAS
  update-lbs             Updates load balancer(s)
  version                Prints version

  Use "bbl [command] --help" for more information about a command.
```

## Known Issues

### Re-running `bbl up` Detaches Instances from GCP LBs

Due to `bbl`'s use of Terraform to create infrastructure on GCP, re-running
`bbl up` after deploying either CloudFoundry or Concourse will detach any
instances that were attached to a load balancer created by `bbl`.

At this time, the only known work-arounds are to recreate the affected instance
using `bosh recreate` **or** use Terraform `0.9.7+`.
