# bosh-bootloader
---

This is a command line utility for standing up a CloudFoundry or Concourse installation 
on an IAAS. This CLI is currently under heavy development, and the initial goal is to 
support bootstrapping a CloudFoundry installation on AWS.

* [CI](https://p-concourse.wings.cf-app.com/teams/system-team-infra-infra1-08f1/pipelines/bosh-bootloader)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## Guides

- [Getting Started on AWS](docs/getting-started-aws.md)
- [Deploying Concourse on GCP](docs/concourse.md)
- [Deploying Cloud Foundry on GCP](docs/cloudfoundry.md)
- [Advanced BOSH Configuration](docs/advanced.md)

## Prerequisites

### Install Dependencies

The following should be installed on your local machine
- Golang >= 1.7 (install with `brew install go`)
- bosh-init ([installation instructions](http://bosh.io/docs/install-bosh-init.html))
- terraform >= 0.8.7 ([download here](https://www.terraform.io/downloads.html))

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
                "iam:*"
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
gcloud iam service-accounts create some-account-name
gcloud iam service-accounts keys create "service-account.key.json" --iam-account "some-account-name@PROJECT_ID.iam.gserviceaccount.com"
gcloud projects add-iam-policy-binding PROJECT_ID --member 'serviceAccount:some-account-name@PROJECT_ID.iam.gserviceaccount.com' --role 'roles/editor'
```

## Usage

The `bbl` command can be invoked on the command line and will display its usage.

```
$ bbl
Usage:
  bbl [GLOBAL OPTIONS] COMMAND [OPTIONS]

Global Options:
  --help      [-h]       Print usage
  --version   [-v]       Print version
  --state-dir            Directory containing bbl-state.json

Commands:
  create-lbs             Attaches load balancer(s)
  delete-lbs             Deletes attached load balancer(s)
  destroy                Tears down BOSH director infrastructure
  director-address       Prints BOSH director address
  director-username      Prints BOSH director username
  director-password      Prints BOSH director password
  director-ca-cert       Prints BOSH director CA certificate
  env-id                 Prints environment ID
  help                   Prints usage
  lbs                    Prints attached load balancer(s)
  ssh-key                Prints SSH private key
  up                     Deploys BOSH director on AWS
  update-lbs             Updates load balancer(s)
  version                Prints version

  Use "bbl [command] --help" for more information about a command.
```
