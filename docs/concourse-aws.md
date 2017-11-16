# Deploying Concourse

This document will walk through deploying a concourse clustered install to AWS using `bbl` and `bosh`.

## Prerequisites

* Install `bbl` CLI
* Completed [BOSH installation to AWS](https://github.com/cloudfoundry/bosh-bootloader/blob/master/docs/getting-started-aws.md)
* [Bosh v2 CLI](https://bosh.io/docs/cli-v2.html) installed

## Create load balancer

First you need to generate self-signed certificates for your domain.

```
openssl req \
       -newkey rsa:2048 -nodes -keyout concourse.example.com.key \
       -x509 -days 365 -out concourse.example.com.crt
```

For now you need to convert RSA key to be usable by `bbl` (see [issue](https://github.com/cloudfoundry/bosh-bootloader/issues/130)):

```
openssl rsa -in concourse.example.com.key -out rsakey.pem
```

Finally, create load balancers and update cloud config:

```
bbl create-lbs \
  --type concourse \
  --cert concourse.eminens.io.crt \
  --key rsakey.pem
```

## Create a bosh deployment manifest

Scale instance types, disks and instance count based on your needs. Other sizes are available, see ```bosh cloud-config```.

1. Start with the sample manifest from the [Concourse documentation](http://concourse.ci/clusters-with-bosh.html)
2. Replace all ```vm_type: REPLACE_ME``` with ```vm_type: t2.small```.
3. Add the vm_extension ```lb``` to the instance_group "web"
4. Delete `tls_cert` and `tls_key` from the properties of the job named `atc`
5. Add the vm_extension ```50GB_ephemeral_disk``` to the instance_group "worker"
6. Replace all ```persistent_disk_type: REPLACE_ME``` with ```persistent_disk_type: 5GB```
7. Replace `director_uuid: REPLACE_ME` with `uuid` from `bosh env`
8. Fill `external_url: REPLACE_ME` with Concourse external URL

## Upload releases

1. Upload latest [stemcell](http://bosh.io/stemcells)
```
bosh upload-stemcell https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent
```
2. Upload latest concourse [BOSH Releases](http://concourse.ci/downloads.html)
```
bosh upload-release https://github.com/concourse/concourse/releases/download/v2.7.3/concourse-2.7.3.tgz
bosh upload-release https://github.com/concourse/concourse/releases/download/v2.7.3/garden-runc-1.4.0.tgz
```
3. Upload latest [postgres release](http://bosh.io/releases/github.com/cloudfoundry/postgres-release?all=1)
```
bosh upload-release https://bosh.io/d/github.com/cloudfoundry/postgres-release
```

## Deploy

```
bosh -d concourse deploy concourse.yml
```

## Verify

Point your browser to `$external_url`.
