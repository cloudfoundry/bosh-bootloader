## <a name='byobastion-gcp'></a>byobastion-gcp

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

