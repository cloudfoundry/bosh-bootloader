# How-To SSH

## To the BOSH director with a jumpbox

Required:
* [bosh-deployment](https://github.com/cloudfoundry/bosh-deployment)

1. Start by applying the jumpbox-user ops-file during bbl up:

    ```
    bbl up --credhub --ops-file bosh-deployment/jumpbox-user.yml
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
    ssh-add jumpbox.key
    ```

1. SSH to the jumpbox:

    ```
    ssh -A jumpbox@<INSERT JUMPBOX EXTERNAL IP> -i $BOSH_GW_PRIVATE_KEY
    ```

1. From the jumpbox, ssh to the director:

    ```
    ssh jumpbox@10.0.0.6
    ```
