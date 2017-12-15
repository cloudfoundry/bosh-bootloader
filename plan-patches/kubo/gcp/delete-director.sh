#!/bin/sh
bosh delete-env \
  ${BBL_STATE_DIR}/bosh-deployment/bosh.yml \
  --state  ${BBL_STATE_DIR}/vars/bosh-state.json \
  --vars-store  ${BBL_STATE_DIR}/vars/director-vars-store.yml \
  --vars-file  ${BBL_STATE_DIR}/vars/director-vars-file.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/gcp/cpi.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/jumpbox-user.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/uaa.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/credhub.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/local-dns.yml \
  -o  ${BBL_STATE_DIR}/bbl-ops-files/gcp/bosh-director-ephemeral-ip-ops.yml 
