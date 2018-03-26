# Plan Patches

Plan patches can be used to customize the IAAS
environment and bosh director that is created by
`bbl up`.

A patch is a directory with a set of files
organized in the same hierarchy as the bbl-state dir.

## bosh-lite-gcp

To create a bosh lite on gcp, the files in `bosh-lite-gcp`
should be copied to your bbl state directory.

The steps might look like such:

```
mkdir banana-env && cd banana-env

bbl plan --name banana-env

cp -r bosh-bootloader/plan-patches/bosh-lite-gcp/. .

bbl up
```


## restricted-instance-groups-gcp

To create two instance groups instead of an instance group for every zone on gcp,
you can use the steps above with the `restricted-instance-groups-gcp` patch
provided here.

## iso-segs-gcp

Creates a single routing isolation segment on GCP, including dedicated load balancers and firewall rules.

```
cp -r bosh-bootloader/plan-patches/iso-segs-gcp/. some-env/
bbl up
```

Disclaimer: this is a testing/development quality patch.  It has not been subject to a security review -- the firewall rules may not be fully locked down.
Please don't run it in production!


## tf-backend-gcp

Stores the terraform state in a bucket in Google Cloud Storage.

```
cp -r bosh-bootloader/plan-patches/tf-backend-gcp/. .
```

Since the backend configuration is loaded by Terraform extremely early (before
the core of Teraform can be initialized), there can be no interplations in the backend
configuration template. Instead of providing vars for the bucket to a `gcs_backend_override.tfvars`,
the values for the bucket name, credential path, and state prefix must be provided directly
in the backend configuration template.

Modify `terraform/gcs_backend_override.tf` to provide the name of the bucket, the path to
the GCP service account key, and a prefix for the environment inside the bucket.

Then you can bbl up.

```
bbl up
```

## byobastion-gcp

To use your own bastion on gcp, the files in `byobastion-gcp`
should be copied to your bbl state directory.

Why would you do this?

You want to use a vm in your vpc that can serve
as the bastion to a bbl'd up director, but that makes
using a bbl'd up jumpbox feel redundant. This patch
will deploy a director that can be accessed directly from the
bastion without a jumpbox, but not from the larger internets.

The steps might look like such:

1. In the gcp console, create a network and a vm, put it in a 10.0.0.0/16 subnet.
1. ssh to that vm, install bbl, terraform, and the bosh cli (and its dependencies.)
1. Create your bbl-state and apply this plan patch
    ```
    mkdir banana-env && cd banana-env

    cp -r bosh-bootloader/plan-patches/byobastion-gcp/. .

    bbl plan --name banana-env
    ```

1. Configure the patch and bbl up:
    ```
    vim vars/bastion.tfvars # fill out variables with the network, subnet, and external ip name you've made in the gcp console

    bbl up # will exit 1
    ```
    Note: `bbl up` fails while trying to update the cloud-config
    because it assumes there is a jumpbox and tries to contact the director
    via an ssh tunnel + proxy through it. To upload a cloud config without proxying
    through your (nonexistant) jumpbox, you can run:

1. Upload your cloud config
    ```
    eval "$(bbl print-env | grep -v BOSH_ALL_PROXY)"

    bosh update-cloud-config cloud-config/cloud-config.yml -o cloud-config/ops.yml
    ```
1. Once your director is deployed, you can target it with:
    ```
    eval "$(bbl print-env | grep -v BOSH_ALL_PROXY)"

    bosh deploy ...
    ```

