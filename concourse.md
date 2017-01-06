# Deploying Concourse
---

This document will walk through deploying a concourse clustered install to GCP using bbl and bosh.

## Prerequisites

* bbl
* A GCP Service Account key as described in README.md
* BBL up, e.g. ```bbl up --gcp-zone us-west1-a --gcp-region us-west1 --gcp-service-account-key service-account.key.json --gcp-project-id my-gcp-project-id --iaas gcp```
* Add a load balancer, e.g. ```bbl create-lbs --type concourse --cert <path-to-tls-cert> --key <path-to-tls-key>```
* This guide will assume the [Bosh v2 CLI](https://bosh.io/docs/cli-v2.html) is installed, but bosh v1 CLI will work, with some minor changes.

## Create load balancer

```
bbl create-lbs --type concourse
```

## Create a bosh deployment manifest

Scale instance types, disks and instance count based on your needs. Other sizes are available, see ```bosh cloud-config```.

1. Start with the sample manifest from the [Concourse documentation](http://concourse.ci/clusters-with-bosh.html)
2. Replace all ```vm_type: REPLACE_ME``` with ```vm_type: n1-standard-2```.
3. Add the vm_extension ```lb``` to the instance_group "web"
4. Add the property ```tls_bind_port: 443``` to the instance_group "web"
5. Add the vm_extension ```50GB_ephemeral_disk``` to the instance_group "worker"
6. Replace all ```persistent_disk_type: REPLACE_ME``` with ```persistent_disk_type: 5GB```
7. Fill out the remaining REPLACE_ME in the sample manifest with your own data, such as auth groups, SSL certs, and external URL


## Set the bosh environment

```
export BOSH_USER=`bbl director-username`
export BOSH_PASSWORD=`bbl director-password`
export BOSH_CA_CERT=`bbl director-ca-cert`
export BOSH_ENVIRONMENT=`bbl director-address`
```

## Upload releases

1. Download and upload latest (Google stemcell)[http://bosh.io/stemcells]
```
bosh upload-stemcell ~/Downloads/light-bosh-stemcell-XXXX.X-google-kvm-ubuntu-trusty-go_agent.tgz
```
2. Download and upload latest concourse (BOSH Releases)[http://concourse.ci/downloads.html]
```
bosh upload-release ~/Downloads/garden-runc-X.X.X.tgz
bosh upload-release ~/Downloads/concourse-2.5.1.tgz
```

## Deploy

```
bosh -n concourse deploy concourse.yml
```
