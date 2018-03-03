# Advanced configuration

## Table of Contents
* <a href='#opsfile'>Using a BOSH ops-file with bbl</a>
* <a href='#terraform'>Customizing IaaS Paving with Terraform</a>
* <a href='#boshlite'>Deploying BOSH lite on GCP</a>
* <a href='#isoseg'>Deploying an isolation segment</a>

## <a name='opsfile'></a>Using a BOSH ops-file with bbl

### About BOSH ops-files

Certain features of BOSH, particularly experimental features or tuning parameters, must be enabled by modifying your
Director's deployment manifest. `bosh-deployment` contains many such [ops files](https://bosh.io/docs/terminology.html#operations-file) for common features and options.

### Using the pre-made operations files
You can provide any number of ops files or variables to `bosh create-env` by creating `create-director-override.sh`. This file will not be overridden by bbl. You can use `create-director.sh` as a template, and you can even edit that file instead, but if you do, your changes will be overridden the next time you run `bbl plan`.

In this example, I use a local version of BOSH director that I have built based off of a branch by referencing an ops file that is included as part of `bosh-deployment`:
```diff
bosh create-env \
  ${BBL_STATE_DIR}/bosh-deployment/bosh.yml \
  --state  ${BBL_STATE_DIR}/vars/bosh-state.json \
  --vars-store  ${BBL_STATE_DIR}/vars/director-vars-store.yml \
  --vars-file  ${BBL_STATE_DIR}/vars/director-vars-file.yml \
+  -o ${BBL_STATE_DIR}/bosh-deployment/local-bosh-release.yml
+  -v local_bosh_release=${BBL_STATE_DIR}/../../build/bosh-dev.tgz
  -o  ${BBL_STATE_DIR}/bosh-deployment/cpi.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/jumpbox-user.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/uaa.yml \
  -o  ${BBL_STATE_DIR}/../shared/bosh-deployment/credhub.yml 
```

### Authoring an ops-file
The [operations files](http://bosh.io/docs/cli-ops-files.html) provided by `bosh-deployment` may not meet your needs. In this case you will have to write your own
custom ops-file. Store it somewhere outside of the bosh-deployment directory. New versions of bbl will keep the
bosh-deployment directory in sync with the latest configuration and releases, so modifications may be lost when
`bbl plan` is run in the future. Consider storing it in the top level of your state directory if it is environmentally
specific, or above the state directory if it is true for all environments.

Here is an example of adding an ops file that configures a few settings on all of my BOSH directors:  
```diff
#!/bin/sh
bosh create-env \
  ${BBL_STATE_DIR}/bosh-deployment/bosh.yml \
  --state  ${BBL_STATE_DIR}/vars/bosh-state.json \
  --vars-store  ${BBL_STATE_DIR}/vars/director-vars-store.yml \
  --vars-file  ${BBL_STATE_DIR}/vars/director-vars-file.yml \
+  -o ${BBL_STATE_DIR}/../../bbl-envs/shared/increase-workers-threads-and-flush-arp.yml
  -o  ${BBL_STATE_DIR}/bosh-deployment/cpi.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/cpi.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/jumpbox-user.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/uaa.yml \
  -o  ${BBL_STATE_DIR}/../shared/bosh-deployment/credhub.yml
```
## <a name='terraform'></a>Customizing IaaS Paving with Terraform
Numerous settings can be reconfigured repeatedly by editing `$BBL_STATE_DIR/vars/terraform.tfvars` or adding a terraform override into  `$BBL_STATE_DIR/terraform/my-cool-template-override.tf`. Some settings, like VPCs, are not able to be changed after initial creation so it may be better to `bbl plan` first before running `bbl up` for the first time.

### Example: adjusting the cidr on AWS
1. Plan the environment:
    ```
    mkdir some-env && cd some-env
    echo BBL_AWS_ACCESS_KEY_ID=MYKEY
    echo BBL_AWS_SECRET_ACCESS_KEY=MYSECRET
    bbl plan --iaas aws --aws-region us-west-1
    echo -e "\nvpc_cidr=\"192.168.0.0/20\"" >> vars/terraform.tfvars
    ```
1. Create the environment:
    ```
    bbl up
    ```
    That's it. Your director is now at `192.168.0.6`.

## <a name='boshlite'></a>Deploying BOSH lite on GCP
1. Plan the environment:
    ```
    git clone https://github.com/cloudfoundry/bosh-bootloader.git
    mkdir some-env && cd some-env
    BBL_GCP_SERVICE_ACCOUNT_KEY=<MYSERVICEACCOUNTKEY>
    bbl plan --name some-env --iaas gcp --gcp-region us-west-1
    cp -r ../bosh-bootloader/plan-patches/bosh-lite-gcp/* .
    ```
1. Create the environment:
    ```
    bbl up
    ```
1. Determine your external IP:
    ```
    bosh int vars/director-vars-file.yml --path /external_ip
    ```
1. Add it to your DNS:
    ```
    bosh-lite.infrastructure.cf-app.com.	A	300	${bosh_lite_external_ip}
    *.bosh-lite.infrastructure.cf-app.com.	CNAME	300	bosh-lite.infrastructure.cf-app.com.
    ```
1. Deploy cf-deployment:
    ```
    $ bosh upload-stemcell https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=3468.5
    $ bosh deploy -d cf -v 'system_domain=cf.evanfarrar.com' -o operations/bosh-lite.yml cf-deployment.yml -o operations/use-compiled-releases.yml
    ```

## <a name='isoseg'></a>Deploying an isolation segment
You can use this process on AWS to create an isolation segment with 
```
mkdir some-env && cd some-env
bbl plan --name some-env --lb-type cf --lb-cert /path/to/lb.crt --lb-key /path/to/lb.key
cp /path/to/patch-dir/cloud-config/iso-segs-ops.yml cloudconfig/
TF_VAR_isolation_segments="1" bbl up
```
To se the TF_VAR it is also possible to add `isolation_segments="1"` to `terraform.tfvars` before running up.
