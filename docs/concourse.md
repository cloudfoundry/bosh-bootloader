# Generic Steps for Concourse Deployment

This document will walk through deploying a concourse clustered
install using `bbl` and `bosh`.

## Prerequisites

- `bbl`
- [bosh v2](https://bosh.io/docs/cli-v2.html)
- [concourse/concourse-bosh-deployment](https://github.com/concourse/concourse-bosh-deployment)

## Steps

1. Create an environment and upload a stemcell.

  ```bash
  bbl up --lb-type concourse

  export external_url="https://$(bbl lbs | awk -F': ' '{print $2}')"

  eval "$(bbl print-env)"

  bosh upload-stemcell https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-trusty-go_agent

  cd $GOPATH/src/github.com/concourse/concourse-bosh-deployment/cluster
  ```

1. Deploy concourse.

  ```bash
  cat >secrets.yml <<EOL
local_user:
    username: <username>
    password: <super-secret-password>
EOL

  bosh deploy -d concourse concourse.yml \
    -l ../versions.yml \
    -l secrets.yml \
    --vars-store cluster-creds.yml \
    -o operations/basic-auth.yml \
    -o operations/privileged-http.yml \
    -o operations/privileged-https.yml \
    -o operations/tls.yml \
    -o operations/web-network-extension.yml \
    --var network_name=default \
    --var external_url=$external_url \
    --var web_vm_type=default \
    --var db_vm_type=default \
    --var db_persistent_disk_type=10GB \
    --var worker_vm_type=default \
    --var deployment_name=concourse \
    --var web_network_name=private \
    --var web_network_vm_extension=lb
  ```

## Verify
Point your browser to `$external_url`.
