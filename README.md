# bosh-bootloader
Also known as `bbl` *(pronounced: "bubble")*, bosh-bootloader is a command line utility for standing up BOSH
on an IaaS. `bbl` currently supports AWS, GCP, Microsoft Azure, Openstack and vSphere.

* [CI](https://infra.ci.cf-app.com/teams/main/pipelines/bosh-bootloader/)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## Docs

- [Getting Started: GCP](docs/getting-started-gcp.md)
- [Deploying Concourse](docs/concourse.md)
- [Upgrade](docs/upgrade.md)
- [Advanced Configuration](docs/advanced-configuration.md)

## Prerequisites

### Install Dependencies

The following should be installed on your local machine
- [bosh-cli](https://bosh.io/docs/cli-v2.html)
- [bosh create-env dependencies](https://bosh.io/docs/cli-env-deps.html)
- [terraform](https://www.terraform.io/downloads.html) >= 0.11.0
- ruby (necessary for bosh create-env)

### Install bosh-bootloader using a package manager

**Mac OS X**

```sh
$ brew tap cloudfoundry/tap
$ brew install bosh-cli
$ brew install bbl
```

## Usage

### IaaS-Specific Getting Started Guides
- [Getting Started: Azure](docs/getting-started-azure.md)
- [Getting Started: GCP](docs/getting-started-gcp.md)
- [Getting Started: AWS](docs/getting-started-aws.md)
- [Getting Started: vSphere](docs/getting-started-vsphere.md)
- [Getting Started: OpenStack](docs/getting-started-openstack.md)

### Managing state

The bbl state directory contains all of the files that were used to create your bosh director. This should be checked in
to version control, so that you have all the information necessary to later destroy or update this environment at a later
date.

 filename |  contents
------------ | -------------
``bbl-state.json`` | Environment name, and bbl version metadata
``terraform/`` | The terraform templates bbl used to pave your IaaS. See [docs/advanced-configuration](docs/advanced-configuration.md#terraform) for information on modifying this. 
``vars/`` | This is where bbl will store environment specific variables. Consider storing this outside of version control.
``jumpbox-deployment/`` | The latest [jumpbox-deployment](http://github.com/cppforlife/jumpbox-deployment) that has been tested with your version of bbl.
``create-jumpbox.sh`` | The BOSH cli command bbl will use to create your jumpbox.
``bosh-deployment/`` | The latest [bosh-deployment](http://github.com/cloudfoundry/bosh-deployment) that has been tested with your version of bbl
``create-director.sh`` | The BOSH cli command bbl will use to create your director when you run `bbl up`. See [docs/advanced-configuration](docs/advanced-configuration.md#opsfile) for help with modifying this.
``cloud-config/``| The cloud-config yaml that bbl will upload to the director to map IAAS resources to BOSH resources.
``delete-director.sh`` | The BOSH cli command bbl will use to delete your director.
``delete-jumpbox.sh`` | The BOSH cli command bbl will use to delete your jumpbox.


### Tearing down an environment

Once you are done kicking the tires on CF and BOSH, clean up your environment to save IaaS costs:

1. You must first delete any deployments on BOSH. e.g. `bosh -d cf delete-deployment`

1. `bbl down` with your IaaS user/account information.

### Automating the automation tool

In order to use `bbl` in your concourse pipelines, the current supported way
for `cf-deployment` is to use the
[cf-deployment-concourse-tasks](https://github.com/cloudfoundry/cf-deployment-concourse-tasks).

There is a work-in-progress concourse resource for bbl:
[bbl-state-resource](https://github.com/cloudfoundry/bbl-state-resource).
