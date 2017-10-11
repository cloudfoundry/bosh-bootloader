# bosh-bootloader
---

This is a command line utility for standing up a CloudFoundry or Concourse installation
on an IAAS. This CLI supports bootstrapping a CloudFoundry or Concourse installation on
AWS and GCP. Azure support is in progress.

* [CI](https://wings.concourse.ci/teams/cf-infrastructure/pipelines/bosh-bootloader)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## Guides

- [AWS - Getting Started](docs/getting-started-aws.md)
- [AWS - Deploying Concourse](docs/concourse-aws.md)
- [GCP - Deploying Concourse](docs/concourse-gcp.md)
- [GCP - Deploying Cloud Foundry](https://github.com/cloudfoundry/cf-deployment/blob/master/deployment-guide.md)
- [Advanced BOSH Configuration](docs/advanced.md)

## Prerequisites

### Install Dependencies

The following should be installed on your local machine
- [bosh-cli](https://bosh.io/docs/cli-v2.html)
- [terraform](https://www.terraform.io/downloads.html) >= 0.10.0
- ruby

### Install bosh-bootloader using a package manager

**Mac OS X**

```sh
$ brew tap cloudfoundry/tap
$ brew install bosh-cli
$ brew install bbl
```

### IAAS Configuration

#### AWS

[Create an IAM user.](docs/getting-started-aws.md#creating-an-iam-user)

#### GCP

Create a service account.

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
  help                    Prints usage
  version                 Prints version
  up                      Deploys BOSH director on an IAAS
  destroy                 Tears down BOSH director infrastructure
  lbs                     Prints attached load balancer(s)
  create-lbs              Attaches load balancer(s)
  update-lbs              Updates load balancer(s)
  delete-lbs              Deletes attached load balancer(s)
  rotate                  Rotates SSH key for the jumpbox user
  bosh-deployment-vars    Prints required variables for BOSH deployment
  jumpbox-deployment-vars Prints required variables for jumpbox deployment
  cloud-config            Prints suggested cloud configuration for BOSH environment
  jumpbox-address         Prints BOSH jumpbox address
  director-address        Prints BOSH director address
  director-username       Prints BOSH director username
  director-password       Prints BOSH director password
  director-ca-cert        Prints BOSH director CA certificate
  env-id                  Prints environment ID
  latest-error            Prints the output from the latest call to terraform
  print-env               Prints BOSH friendly environment variables
  ssh-key                 Prints SSH private key

  Use "bbl [command] --help" for more information about a command.
```

### Generic steps to a Cloud Foundry deployment

Step 1: Create the necessary IAAS user/account for bbl.
Step 2: `bbl up` with IAAS credentials as flags or environment variables.
Step 3: `bbl create-lbs --type cf` with a certificate and key as flags or environment variables. (Continue to provide IAAS credentials as flags or environment variables.)
Step 4: `eval "$(bbl print-env)"` to create an SSH tunnel to the BOSH director for Step 5.
Step 5: `bosh deploy` with the path to the manifest you intend to deploy!

To tear down load balancers, run `bbl delete-lbs`.
To tear it all down, run `bbl destroy`.

Note: You must delete your BOSH deployments before running `bbl destroy`.
