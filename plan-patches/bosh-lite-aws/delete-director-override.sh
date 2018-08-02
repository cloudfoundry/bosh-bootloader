#!/usr/bin/env bash
bosh delete-env
 ${BBL_STATE_DIR}/bosh-deployment/bosh.yml \
 --state=state.json \
 --vars-store  ${BBL_STATE_DIR}/vars/director-vars-store.yml \
 --vars-file  ${BBL_STATE_DIR}/vars/director-vars-file.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/aws/cpi.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/bosh-lite.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/bosh-lite-runc.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/jumpbox-user.yml \
 -o ${BBL_STATE_DIR}/bbl-ops-files/aws/bosh-director-ephemeral-ip-ops.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/aws/iam-instance-profile.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/uaa.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/credhub.yml \
 -v access_key_id="${BBL_AWS_ACCESS_KEY_ID}" \
 -v secret_access_key="${BBL_AWS_SECRET_ACCESS_KEY}"
