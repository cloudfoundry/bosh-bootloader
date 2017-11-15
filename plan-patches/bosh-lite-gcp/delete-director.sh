#!/bin/sh
bosh delete-env \
  ${BBL_STATE_DIR}/bosh-deployment/bosh.yml \
  --state  ${BBL_STATE_DIR}/vars/bosh-state.json \
  --vars-store  ${BBL_STATE_DIR}/vars/director-variables.yml \
  --vars-file  ${BBL_STATE_DIR}/vars/director-deployment-vars.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/gcp/cpi.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/jumpbox-user.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/uaa.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/credhub.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/bosh-lite.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/bosh-lite-runc.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/gcp/bosh-lite-vm-type.yml \
  -o  ${BBL_STATE_DIR}/external-ip-gcp.yml
