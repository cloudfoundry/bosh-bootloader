# How-To SSH

## To the BOSH director

1. Set up a SOCKS5 proxy by running:

    ```
    eval "$(bbl print-env)"
    ```

1. Interpolate out the jumpbox user's ssh key for reaching the director:

    ```
    bbl director-ssh-key > director-jumpbox-user.key
    chmod 600 director-jumpbox-user.key
    ```

1. SSH via the proxy:

    ```
    ssh -o ProxyCommand='nc -x localhost:`echo $BOSH_ALL_PROXY| cut -f3 -d':'` %h %p' \
        -i /tmp/director-jumpbox-user.key jumpbox@10.0.0.6
    ```

## To job VMs

The command 
```
eval "$(bbl print-env)"
bosh ssh web/0
```
