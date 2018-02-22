# Howto ssh

## To the jumpbox

1. Use print-env to see the ssh command:

    ```
    bbl print-env
    ```

1. Remove the `-f -N` and `-D PORT`, then run it:

    ```
    ssh -o StrictHostKeyChecking=no -o ServerAliveInterval=300 \
        jumpbox@34.214.217.33 -i $JUMPBOX_PRIVATE_KEY
    ```

## To the BOSH director

1. Set up a SOCKS5 proxy by running:

    ```
    eval "$(bbl print-env)"
    ```

1. Interpolate out the jumpbox user's ssh key for reaching the director:

    ```
    bbl director-ssh-key > /tmp/director-jumpbox-user.key
    chmod 600 /tmp/director-jumpbox-user.key
    ```

1. SSH via the proxy:

    ```
    ssh -o ProxyCommand='nc -x localhost:`echo $BOSH_ALL_PROXY| cut -f3 -d':'` %h %p' \
        -i /tmp/director-jumpbox-user.key jumpbox@10.0.0.6
    ```

## To job VMs

The command `print-env` will print out everything necessary to ssh to a job VM (including a SOCKS5 proxy to the director's private network via ).
```
eval "$(bbl print-env)"
bosh ssh web/0
```

### Troubleshooting
* It is not necessary to set BOSH_GW_HOST and other old-style `bosh ssh` variables. Unset them.
* The ubuntu stemcell allows a maximum of three login attempts, so ensure you do not have a lot of keys in your SSH keyring. `ssh-add -D` can clear them all.
