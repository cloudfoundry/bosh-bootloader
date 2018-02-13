# Accessing the BOSH director credhub

## Using CREDHUB_PROXY

### requirements

- `bbl v6.2`
- `credhub-cli v1.6.0`
- a bbl environment

### steps

```
eval "$(bbl print-env)"

credhub find -n 'cf_admin_password'
```


## Using http_proxy

Using `credhub-cli < v1.6.0`.

### steps

1. Set your CredHub client/secret

    ```
    eval "$(bbl print-env)"
    ```

1. Make an SSH tunnel to the jumpbox

    ```
    bbl ssh-key > /tmp/jumpbox.key
    chmod 0700 /tmp/jumpbox.key
    ssh -4 -D 5000 -fNC jumpbox@`bbl jumpbox-address` -i /tmp/jumpbox.key
    ```

1. Login
    ```
    http_proxy=socks5://localhost:5000 credhub login
    ```

1. Get credentials
    ```
    http_proxy=socks5://localhost:5000 credhub find -n 'cf_admin_password'
    ```
