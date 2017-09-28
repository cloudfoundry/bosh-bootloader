# How-To SSH

## To the BOSH director if you have a jumpbox (v5.0.0+)

Required:
* [bosh-deployment](https://github.com/cloudfoundry/bosh-deployment)

1. Set JUMPBOX_PRIVATE_KEY by running:

    ```
    eval "$(bbl print-env)"
    ```

1. Retrieve the bosh credentials from the bbl-state:

    ```
    cat bbl-state.json | jq --raw-output .bosh.variables > creds.yml
    ```

1. Interpolate out the jumpbox user's ssh key for reaching the director:

    ```
    bosh int creds.yml --path /jumpbox_ssh/private_key > jumpbox-user.key
    chmod 600 jumpbox-user.key
    ```

1. Add the key to the SSH agent:

    ```
    ssh-add jumpbox-user.key
    ```

1. SSH to the jumpbox:

    ```
    ssh -A jumpbox@`bbl jumpbox-address|sed 's/:22//'` -i $JUMPBOX_PRIVATE_KEY
    ```

1. From the jumpbox, ssh to the director:

    ```
    ssh jumpbox@10.0.0.6
    ```
