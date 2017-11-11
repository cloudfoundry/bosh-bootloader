# Deploying Concourse

This document will walk through deploying a concourse clustered install to GCP using bbl and bosh.

## Pre-requisites
* bbl
* terraform
* a git repo for storing state
* [A GCP service account key](getting-started-gcp.md#creating-a-service-account) in `~/service-account-key.json`

## Plan a new environment
* make a new directory in your environments repo
  ```
  $ cd environments
  $ mkdir my-concourse
  $ cd my-concourse
  ```
* initialize the state directory
  ```
  export BBL_GCP_SERVICE_ACCOUNT_KEY=~/service-account-key.json
  $ bbl plan --iaas gcp \
             --name my-concourse \
             --gcp-region us-west1 \
             --lb-type concourse
  ```
* You can also make other [customizations](advanced.md) to the terraform and BOSH director before continuing.

## Create the director

  ```
  bbl up
  ```

## Target the director

  ```
  eval "$(bbl print-env)"
  ```

## Download concourse-deployment
Download the latest release of [concourse-deployment](https://github.com/evanfarrar/concourse-deployment/releases/latest).
  ```
  mv ~/Dowloads/concourse-deployment-0.0.1.tar.gz ./
  tar xvf concourse-deployment-0.0.1.tar.gz
  ```
## Target the director
  ```
  eval "$(bbl print-env)"
  ```

## Upload A Stemcell

* Download and upload latest [Google stemcell](http://bosh.io/stemcells)
  ```
  bosh upload-stemcell ~/Downloads/light-bosh-stemcell-XXXX.X-google-kvm-ubuntu-trusty-go_agent.tgz
  ```

## Deploy
  ```
  bosh deploy    \
    -d concourse \
    -v 'system_domain=myconcourse.example.com' \
    -o concourse-deployment-0.0.1/operations/gcp.yml  \
    concourse-deployment-0.0.1/concourse-deployment.yml
  ```
  
## Set up DNS entry
Run `bbl lbs` to discover the IP address of your load balancer. Then set up a A record in your DNS provider to point `myconcourse.example.com` to this IP.

## Visit your concourse UI
* Open http://myconcourse.example.com and download the `fly` CLI for your platform

## Retrieve passwords
* Get your CredHub password
```
bosh interpolate vars/director-variables.yml --path /credhub_cli_password
```
* SSH to the jumpbox
```
ssh -i $JUMPBOX_PRIVATE_KEY -t jumpbox@`../bbl-v5.4.0_osx jumpbox-address` bash
```

* Download the credhub CLI to the jumpbox
```
wget https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/1.5.0/credhub-linux-1.5.0.tgz
tar xvf credhub-linux-1.5.0.tgz
exit
```
* retrieve the CredHub CA, UAA CA, and CredHub UAA User Password from BOSH director Var Store in BBL State
```
export CREDHUB_CLI_PASSWORD=`bosh interpolate vars/director-variables.yml --path /credhub_cli_password`
export UAA_CA="$(bosh interpolate vars/director-variables.yml --path /uaa_ssl/ca)"
export CREDHUB_CA="$(bosh interpolate vars/director-variables.yml --path /credhub_ca/ca)"
```
* retrieve the concourse username and password from credhub
```
export CONCOURSE_USERNAME=$(ssh -i $JUMPBOX_PRIVATE_KEY -t jumpbox@`../bbl-v5.4.0_osx jumpbox-address` "export CREDHUB_CA=\"$CREDHUB_CA\"; export UAA_CA=\"$UAA_CA\"; cd ~; ./credhub login -s https://10.0.0.6:8844 --ca-cert \"\$UAA_CA\" --ca-cert \"\$CREDHUB_CA\" -u credhub-cli -p $CREDHUB_CLI_PASSWORD>/dev/null; ./credhub get --name \`./credhub find -n 'basic_auth_username' | grep -e'- name' | cut -d':' -f2\` | grep value | cut -d":" -f2 | tr -d '[:space:]'" 2>/dev/null)
export CONCOURSE_PASSWORD=$(ssh -i $JUMPBOX_PRIVATE_KEY -t jumpbox@`../bbl-v5.4.0_osx jumpbox-address` "export CREDHUB_CA=\"$CREDHUB_CA\"; export UAA_CA=\"$UAA_CA\"; cd ~; ./credhub login -s https://10.0.0.6:8844 --ca-cert \"\$UAA_CA\" --ca-cert \"\$CREDHUB_CA\" -u credhub-cli -p $CREDHUB_CLI_PASSWORD>/dev/null; ./credhub get --name \`./credhub find -n 'basic_auth_password' | grep -e'- name' | cut -d':' -f2\` | grep value | cut -d":" -f2 | tr -d '[:space:]'" 2>/dev/null)
```

Now you should be able to login with `fly`* :
```
~/Downloads/fly -t concourse login -c https://concourse.evanfarrar.com/ -u $CONCOURSE_USERNAME -p $CONCOURSE_PASSWORD
```
(* depending on your certs and domain validity, you may need to supply `-k`. TODO: get concourse CA from credhub.)

Enjoy!
