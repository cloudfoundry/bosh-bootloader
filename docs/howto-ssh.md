# How To SSH

## To the Jumpbox

This command shells out to `ssh` to initiate an interactive ssh session to the jumpbox vm.
```
bbl ssh --jumpbox
```

## To the BOSH Director

This command will shell out to `ssh` twice. On the first invocation, it will open a tunnel forwarding a random port to the jumpbox.

On the second invocation, it initiates an interactive ssh session through that port to ssh to the director.

```
bbl ssh --director
```

## To BOSH-Deployed VMs

`bbl print-env` prints out environment variables (`BOSH_ALL_PROXY`, `BOSH_CLIENT`, `BOSH_CLIENT_SECRET`, and others)
that need to be exported to `bosh ssh` to a job vm using the bosh-cli.

Evaluating the command output sets those variables in your environment.

```
bbl print-env | eval
bosh ssh web/0
```

When you run `bosh ssh web/0`, the following happens:

1. The bosh-cli parses `BOSH_ALL_PROXY` and determines from the `ssh+socks5://` scheme that it should proxy through a jumpbox via a tunnel of its own making.

1. The bosh-cli uses some go libraries to start a socks5 proxy on another goroutine. This socks5 proxy is backed by an ssh tunnel from your local machine to the jumpbox.

1. The bosh-cli uses your system's openssh `ssh` "ProxyCommand" option and bsd `nc -x` to open an additional tunnel to `web/0` through that socks5 proxy.

1. When `ssh` exits after you ctrl-D or your ttyless command exits, the bosh-cli exits and the socks5 proxy stops with it.

For http requests to the bosh director, the bosh-cli reads `BOSH_ALL_PROXY=ssh+socks5://`
and uses golang's `ssh.Client.Dial` in the cli's http.Client to send each http request
to the director through an ssh tunnel between your local machine and the jumpbox.

### Troubleshooting

* It is not necessary to set BOSH_GW_HOST and other old-style `bosh ssh` variables. Unset them.

* The ubuntu stemcell allows a maximum of three login attempts, so ensure you do not have a lot of keys in your SSH keyring. `ssh-add -D` can clear them all.
