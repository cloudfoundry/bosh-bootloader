# Deploying Cloud Foundry
---

This document will walk through deploying a cf-deployment based Cloud Foundry.

## Prerequisites

* bbl
* A GCP Service Account key as described in README.md
* BBL up. e.g. ```bbl up --gcp-zone us-west1-a --gcp-region us-west1 --gcp-service-account-key service-account.key.json --gcp-project-id my-gcp-project-id --iaas gcp```
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
2. Replace static IP addresses defined in ```cf-deployment.yml``` and ```opsfiles/gcp.yml``` with ip addresses from the cloud-config for bbl:
```
- type: replace
  path: /instance_groups/name=mysql/networks/name=private/static_ips
  value:
-  - &mysql_ip 10.0.1.193
+  - &mysql_ip 10.0.31.190
- type: replace
  path: /instance_groups/name=consul/networks/name=private/static_ips
  value: &consul_ips
-  - 10.0.1.194
-  - 10.0.1.195
-  - 10.0.1.196
+  - 10.0.31.194
+  - 10.0.47.190
+  - 10.0.47.194
- type: replace
  path: /instance_groups/name=nats/networks/name=private/static_ips
  value: &nats_ips
-  - 10.0.1.197
-  - 10.0.1.198
+  - 10.0.31.195
+  - 10.0.31.196
```

3. Replace the ```vm_extensions:``` property of ```diego-brain``` with an empty array in cf-deployment.yml:
```
-  vm_extensions:
-  - ssh-proxy-lb
+  vm_extensions: []
```
4. Remove zone names that do not exist in your region. For example, ```us-west1``` used above contains only two zones, so z3 must be removed from the manifest in two places:
```
instance_groups:
   azs:
   - z1
   - z2
-  - z3
```
6. Create a vars file, cf-deployment-vars.yml:
```
system_domain: cf.example.com
```
7. Check that the rest of the required credentials can autogenerate:
```
bosh -n interpolate --vars-store cf-deployment-vars.yml -o opsfiles/gcp.yml -o opsfiles/disable-router-tls-termination.yml --var-errs cf-deployment.yml
```

## Upload Stemcells

1. Download and upload latest [Google stemcell](http://bosh.io/stemcells)
```
bosh upload-stemcell ~/Downloads/light-bosh-stemcell-XXXX.X-google-kvm-ubuntu-trusty-go_agent.tgz
```

## Deploy

```
bosh -d cf deploy cf-deployment.yml
```
