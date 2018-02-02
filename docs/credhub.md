# Accessing the BOSH director credhub

## Retrieve passwords
* Set your CredHub client/secret
```
eval "$(bbl print-env)"
```
## Make an SSH tunnel to the jumpbox
```
$ bbl ssh-key > /tmp/jumpbox.key
$ chmod 0700 /tmp/jumpbox.key
ssh -4 -D 5000 -fNC jumpbox@`bbl jumpbox-address` -i /tmp/jumpbox.key
```
## Log in
```
http_proxy=socks5://localhost:5000 credhub login
```
## Do stuff
```
http_proxy=socks5://localhost:5000 credhub find -n 'cf_admin_password'
```
