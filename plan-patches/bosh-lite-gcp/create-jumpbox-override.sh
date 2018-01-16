#!/bin/sh
bosh create-env \
  ${BBL_STATE_DIR}/jumpbox-deployment/jumpbox.yml \
  --state  ${BBL_STATE_DIR}/vars/jumpbox-state.json \
  --vars-store  ${BBL_STATE_DIR}/vars/jumpbox-vars-store.yml \
  --vars-file  ${BBL_STATE_DIR}/vars/jumpbox-vars-file.yml \
  --var-file gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}" \
  -v project_id="${BBL_GCP_PROJECT_ID}" \
  -v zone="${BBL_GCP_ZONE}" \
  -o  ${BBL_STATE_DIR}/jumpbox-deployment/gcp/cpi.yml \
  -o ${BBL_STATE_DIR}/ip-forwarding.yml
