# Accessing the BOSH Director CredHub

## Using `CREDHUB_PROXY`

### Requirements

- `bbl v6.2` or above
- `credhub-cli v1.6.0` or above
- a bbl environment

### Instructions

1. Set necessary environment variables

    `bbl print-env` prints out environment variables (`CREDHUB_CLIENT`, `CREDHUB_SECRET`, `CREDHUB_PROXY`,
`CREDHUB_SERVER`, `CREDHUB_CA_CERT`, and others) that need to be exported to target the Director CredHub using the CredHub CLI.

    ```
    eval "$(bbl print-env)"
    ```

1. Get credentials

    ```
    credhub find -n 'cf_admin_password'
    ```

    The CredHub CLI will parse `CREDHUB_PROXY` and determines from the `ssh+socks5://` scheme that it should proxy throuhg a jumpbox via a tunnel of its own making.

## Using `http_proxy`

### Requirements

- `bbl`
- `credhub-cli`
- a bbl environment

### Instructions

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

