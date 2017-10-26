# bosh-bootloader
bbl (pronounced: "bubble") is a command line utility for standing up a CloudFoundry or Concourse installation
on an IAAS. This CLI supports bootstrapping a CloudFoundry or Concourse installation on
AWS, GCP and Azure. Openstack and vSphere support are planned.

* [CI](https://wings.concourse.ci/teams/cf-infrastructure/pipelines/bosh-bootloader)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## Guides

- [AWS - Getting Started](docs/getting-started-aws.md)
- [AWS - Deploying Concourse](docs/concourse-aws.md)
- [GCP - Deploying Concourse](docs/concourse-gcp.md)
- [GCP - Deploying Cloud Foundry](deployment-guide.md)
- [Advanced BOSH Configuration](docs/advanced.md)

## Prerequisites

### Install Dependencies

The following should be installed on your local machine
- [bosh-cli](https://bosh.io/docs/cli-v2.html)
- [terraform](https://www.terraform.io/downloads.html) >= 0.10.0
- ruby (necessary for bosh create-env)

### Install bosh-bootloader using a package manager

**Mac OS X**

```sh
$ brew tap cloudfoundry/tap
$ brew install bosh-cli
$ brew install bbl
```

## Usage

### Generic getting started guide

1. Create the necessary IAAS user/account for bbl.

1. `bbl up --iaas <MY IAAS>` with IAAS credentials as flags or environment variables.

1. `eval $(bbl print-env)` to target the director that you just created.

1. `bosh ssh` `bosh deploy` or `bosh status` should all just work with no further information needed from bbl. 


For detailed, IAAS specific instructions see your preferred IAAS's getting started guide:
- [Azure - Getting Started](docs/getting-started-azure.md)
- [GCP - Getting Started](docs/getting-started-gcp.md#creating-a-service-account)
- [AWS - Getting Started](docs/getting-started-aws.md#creating-an-iam-user)

### Generic steps for Cloud Foundry deployment

1. Create an environment and target the BOSH director as described above

1. `bbl create-lbs --type cf` with a certificate and key as flags or environment variables.
(Continue to provide the IAAS credentials as flags or environment variables.)

1. `bosh deploy cf.yml -o operations/<MY IAAS>` using the [CF deployment manifest!](https://github.com/cloudfoundry/cf-deployment)

### Managing state

The bbl state directory contains all of the files that were used to create your bosh director. This should be checked in
to version control, so that you have all the information necessary to later destroy or update this environment at a later
date.

 filename |  contents
------------ | -------------
``bbl-state.json`` | Environment name, and bbl version metadata
``bosh-deployment`` | The latest [bosh-deployment](http://github.com/cloudfoundry/bosh-deployment) that has been tested with your version of bbl
``create-director.sh`` | The BOSH cli command bbl will use to create your director when you run `bbl up`. See [docs/advanced]](docs/advanced.md#opsfile) for help with modifying this
``create-jumpbox.sh`` | The BOSH cli command bbl will use to create your jumpbox.
``delete-director.sh`` |The BOSH cli command bbl will use to delete your director.
``delete-jumpbox.sh`` | The BOSH cli command bbl will use to delete your jumpbox.
``jumpbox-deployment`` | The latest [jumpbox-deployment](http://github.com/cppforlife/jumpbox-deployment) that has been tested with your version of bbl.
``terraform`` | The terraform templates bbl used to pave your IaaS. See [docs/advanced#terraform]() for information on modifying this.
``vars `` | This is where bbl will store environment specific variables. Consider storing this outside of version control.

### Tearing down an environment

Once you are done kicking the tires on CF and BOSH, clean up your environment to save IAAS costs:

1. You must first delete any deployments on BOSH. e.g. `bosh -d cf delete-deployment`

1. `bbl down` with your IAAS user/account information.