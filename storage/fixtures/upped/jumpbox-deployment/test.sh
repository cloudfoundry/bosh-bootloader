#!/bin/bash

set -eu

vars_store_prefix=/tmp/jumpbox-deployment-test

clean_tmp() {
  rm -f ${vars_store_prefix}.*
}

trap clean_tmp EXIT

# Only used for tests below. Ignore it.
function bosh() {
  shift 1
  command bosh int --var-errs --var-errs-unused ${@//--state=*/} > /dev/null
}

echo "- AWS"
bosh create-env jumpbox.yml \
  -o aws/cpi.yml \
  --vars-store $(mktemp ${vars_store_prefix}.XXXXXX) \
  -v internal_cidr=test \
  -v internal_gw=test \
  -v internal_ip=test \
  -v external_ip=test \
  -v access_key_id=test \
  -v secret_access_key=test \
  -v az=test \
  -v region=test \
  -v default_key_name=test \
  -v default_security_groups=[test] \
  -v private_key=test \
  -v subnet_id=test \
  > /dev/null

echo "- GCP"
bosh create-env jumpbox.yml \
  -o gcp/cpi.yml \
  --vars-store $(mktemp ${vars_store_prefix}.XXXXXX) \
  -v internal_cidr=test \
  -v internal_gw=test \
  -v internal_ip=test \
  -v external_ip=test \
  -v gcp_credentials_json=test \
  -v project_id=test \
  -v zone=test \
  -v tags=[test] \
  -v network=test \
  -v subnetwork=test \
  > /dev/null

echo "- Openstack"
bosh create-env jumpbox.yml \
  -o openstack/cpi.yml \
  --vars-store $(mktemp ${vars_store_prefix}.XXXXXX) \
  -v internal_cidr=test \
  -v internal_gw=test \
  -v internal_ip=test \
  -v external_ip=test \
  -v auth_url=test \
  -v az=test \
  -v default_key_name=test \
  -v default_security_groups=test \
  -v net_id=test \
  -v openstack_password=test \
  -v openstack_username=test \
  -v private_key=test \
  -v region=test \
  -v tenant=test \
  > /dev/null

echo "- vSphere"
bosh create-env jumpbox.yml \
  -o vsphere/cpi.yml \
  -o no-external-ip.yml \
  --vars-store $(mktemp ${vars_store_prefix}.XXXXXX) \
  -v internal_cidr=test \
  -v internal_gw=test \
  -v internal_ip=test \
  -v network_name=test \
  -v vcenter_dc=test \
  -v vcenter_ds=test \
  -v vcenter_ip=test \
  -v vcenter_user=test \
  -v vcenter_password=test \
  -v vcenter_templates=test \
  -v vcenter_vms=test \
  -v vcenter_disks=test \
  -v vcenter_cluster=test \
  > /dev/null
