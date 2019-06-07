# Advanced configuration

## Table of Contents
* <a href='#opsfile'>Using a BOSH ops-file with bbl</a>
* <a href='#terraform'>Customizing IaaS Paving with Terraform</a>
* <a href='#plan-patches'>Applying and authoring plan patches, bundled modifications to default bbl configurations.</a>

## <a name='opsfile'></a>Using a BOSH ops-file with bbl

### About BOSH ops-files

Certain features of BOSH, particularly experimental features or tuning parameters, must be enabled by modifying your
Director's deployment manifest. [`bosh-deployment`](https://github.com/cloudfoundry/bosh-deployment) contains many such [ops files](https://bosh.io/docs/terminology.html#operations-file) for common features and options.

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
  -o  ${BBL_STATE_DIR}/bosh-deployment/jumpbox-user.yml \
  -o  ${BBL_STATE_DIR}/bosh-deployment/uaa.yml \
  -o  ${BBL_STATE_DIR}/../shared/bosh-deployment/credhub.yml
```
## <a name='terraform'></a>Customizing IaaS Paving with Terraform
Numerous settings can be reconfigured repeatedly by editing `$BBL_STATE_DIR/vars/bbl.tfvars` or adding a terraform override into  `$BBL_STATE_DIR/terraform/my-cool-template-override.tf`. Some settings, like VPCs, are not able to be changed after initial creation so it may be better to `bbl plan` first before running `bbl up` for the first time.

### Example: adjusting the cidr on AWS
1. Plan the environment:
    ```
    mkdir some-env && cd some-env
    export BBL_IAAS=aws
    export BBL_AWS_REGION=us-west-1
    export BBL_AWS_ACCESS_KEY_ID=12345678
    export BBL_AWS_SECRET_ACCESS_KEY=12345678
    bbl plan
    echo -e "\nvpc_cidr=\"192.168.0.0/20\"" >> vars/bbl.tfvars
    ```
1. Create the environment:
    ```
    bbl up
    ```
    That's it. Your director is now at `192.168.0.6`.

## <a name='plan-patches'> [Plan Patches](https://github.com/cloudfoundry/bosh-bootloader/tree/master/plan-patches)

Through operations files and terraform overrides, all sorts of wild modifications can be done to the vanilla bosh environments that bbl creates. The basic principal of a plan patch is to make several modifications to a bbl plan in override files that bbl finds under `terraform/`, `cloud-config/`, and `{create,delete}-{jumpbox,director}.sh` . BBL will read and merge those into it's plan when you run `bbl up`.

We've used plan patches to [deploy bosh-lite directors on gcp](https://github.com/cloudfoundry/bosh-bootloader/tree/master/plan-patches/bosh-lite-gcp), to deploy CF Isolation Segments on [public](https://github.com/cloudfoundry/bosh-bootloader/tree/master/plan-patches/iso-segs-gcp) [clouds](https://github.com/cloudfoundry/bosh-bootloader/tree/master/plan-patches/iso-segs-aws), and to deploy bosh managed k8s clusters with working cloud-providers using [cfcr](https://github.com/cloudfoundry-incubator/kubo-deployment/tree/master/manifests).

Our plan patches are experimental. They were tested a bit when we wrote them, but we don't continuously integrate against their dependencies or even check if they still work with recent versions of terraform. They should be used with caution. Operators should make sure they understand each modification and its implications before using our patches in their own environments. Regardless, the plan-patches in this repo are great examples of the different ways you can configure bbl to deploy whatever you might need. To see all the plan patches, visit the [Plan Patches README.md](https://github.com/cloudfoundry/bosh-bootloader/tree/master/plan-patches). If you write your own plan patch that gets you what you need, please consider upstreaming it in a PR.
