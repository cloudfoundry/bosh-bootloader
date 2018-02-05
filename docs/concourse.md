# Deploying Concourse

This document will walk through deploying a concourse clustered
install to AWS, GCP and Azure using `bbl` and `bosh`.

## Prerequisites

* Install `bbl` CLI
* Completed [BOSH installation to AWS](https://github.com/cloudfoundry/bosh-bootloader/blob/master/docs/getting-started-aws.md)
* [Bosh v2 CLI](https://bosh.io/docs/cli-v2.html) installed

## Create load balancer

```bash
bbl plan --lb-type concourse
export external_url=`bbl lbs |awk -F':' '{print $2}' |sed 's/ //'`
# on gcp, you'll need to add http:// to this
```


# Upload latest stemcell
```bash
bosh upload-stemcell https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent
```

## Make an ops file
```bash
cd concourse-deployment/cluster
cat > bbl_ops.yml << 'EOF'
- type: replace
  path: /instance_groups/name=web/vm_extensions?/-
  value: lb
- type: replace
  path: /instance_groups/name=web/jobs/name=atc/properties/bind_port?
  value: 80
EOF
```

## Deploy concourse-deployment

```bash
bosh deploy -d concourse concourse.yml \
  -l ../versions.yml \
  --vars-store cluster-creds.yml \
  -o operations/no-auth.yml \
  -o bbl_ops.yml \
  --var network_name=default \
  --var external_url=$external_url \
  --var web_vm_type=default \
  --var db_vm_type=default \
  --var db_persistent_disk_type=10GB \
  --var worker_vm_type=default \
  --var deployment_name=concourse
```

## Verify
Point your browser to `$external_url`.
