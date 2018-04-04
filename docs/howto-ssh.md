# How To SSH

## To the Jumpbox

```
bbl ssh --jumpbox
```

## To the BOSH Director

```
bbl ssh --director
```

## To BOSH-Deployed VMs

The command `print-env` will print out everything necessary to ssh to a job VM (including a SOCKS5 proxy to the director's private network).

Evaluating the command exports sets those variables in your environment in order to run `bosh ssh`.

```
eval "$(bbl print-env)"
bosh ssh web/0
```

### Troubleshooting

* It is not necessary to set BOSH_GW_HOST and other old-style `bosh ssh` variables. Unset them.
* The ubuntu stemcell allows a maximum of three login attempts, so ensure you do not have a lot of keys in your SSH keyring. `ssh-add -D` can clear them all.
