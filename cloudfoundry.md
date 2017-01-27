# Deploying Cloud Foundry
---

This document will walk through deploying a cf-deployment based Cloud Foundry.

## Prerequisites

* bbl
* A GCP Service Account key as described in README.md
* BBL up. e.g. ```bbl up --gcp-zone us-east1-b --gcp-region us-east1 --gcp-service-account-key service-account.key.json --gcp-project-id my-gcp-project-id --iaas gcp```
* This guide will assume the [Bosh v2 CLI](https://bosh.io/docs/cli-v2.html) is installed, but bosh v1 CLI could work, with some minor changes.

## Set the bosh environment

```
export BOSH_USER=`bbl director-username`
export BOSH_PASSWORD=`bbl director-password`
export BOSH_CA_CERT=`bbl director-ca-cert`
export BOSH_ENVIRONMENT=`bbl director-address`
```

## Create load balancer

```
bbl create-lbs --type cf --key path/to/cf.example.com.key --cert path/to/cf.example.com.crt

```

## Create a bosh deployment manifest

Scale instance types, disks and instance count based on your needs. Other sizes are available, see ```bosh cloud-config```.

1. Start with a clone of the [CF Deployment](https://github.com/cloudfoundry/cf-deployment) repo.
2. Create a vars file, cf-deployment-vars.yml:
```
system_domain: cf.example.com
```
3. Check that the rest of the required credentials can autogenerate:
```
bosh -n interpolate --vars-store cf-deployment-vars.yml -o operations/gcp.yml -o operations/disable-router-tls-termination.yml --var-errs cf-deployment.yml
```

## Upload Stemcells

1. Download and upload latest [Google stemcell](http://bosh.io/stemcells)
```
bosh upload-stemcell ~/Downloads/light-bosh-stemcell-XXXX.X-google-kvm-ubuntu-trusty-go_agent.tgz
```

## Deploy

```
bosh -d cf deploy --vars-store cf-deployment-vars.yml -o opsfiles/gcp.yml cf-deployment.yml

```
