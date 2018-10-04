# jumpbox-deployment

Deploy single vanilla jumpbox machine. Works well with BOSH CLI SOCKS5 proxying.

IMPORTANT: Make sure to configure security group to allow only necessary traffic! Better yet drop all incoming traffic when jumpbox is not being used.

## Planned

- Apply iptables rule to block all incoming traffic
  - in addition to relying on IaaS security groups configuration
- Stop all software aside from SSH after deploy is finished
- Add `--vars-store /dev/null` CLI support?

## Example on AWS

Requires new [BOSH CLI v0.0.146+](https://github.com/cloudfoundry/bosh-cli).

```
$ git clone https://github.com/cppforlife/jumpbox-deployment ~/jumpbox-deployment

$ mkdir -p ~/deployments/jumpbox-1

$ cd ~/deployments/jumpbox-1

# Deploy a jumpbox -- ./creds.yml is generated automatically
$ bosh create-env ~/jumpbox-deployment/jumpbox.yml \
  --state ./state.json \
  -o ~/jumpbox-deployment/aws/cpi.yml \
  --vars-store ./creds.yml \
  -v access_key_id=... \
  -v secret_access_key=... \
  -v region=us-east-1 \
  -v az=us-east-1b \
  -v default_key_name=jumpbox \
  -v default_security_groups=[jumpbox] \
  -v subnet_id=subnet-... \
  -v internal_cidr=10.0.0.0/24 \
  -v internal_gw=10.0.0.1 \
  -v internal_ip=10.0.0.5 \
  -v external_ip=... \
  --var-file private_key=...

# Currently, none of the generated credentials are necessary to persist
# (possibly except for generated SSH private key)
$ rm ./creds.yml
```

Above command requires only two ports open:

```
Type            Protocol Port Range  Source          Purpose
SSH             TCP      22          <BOSH CLI's IP> SSH for bootstrapping & final access
Custom TCP Rule TCP      6868        <BOSH CLI's IP> Agent for bootstrapping
```

## SSH into jumpbox

By default `jumpbox` user is added via `user_add` job. Unique SSH private key is generated.

```
$ bosh int ./creds.yml --path /jumpbox_ssh/private_key > jumpbox.key && chmod 600 jumpbox.key

$ ssh jumpbox@... -i jumpbox.key
```

## Consider using SOCKS5 proxying

Instead of running CLI *from* the jumpbox VM, you can use it as a proxy.

```
# Start SOCKS5 proxy on your machine
$ ssh -N -D 9999 jumpbox@... -i jumpbox.key -f

# Let CLI know about it
$ export BOSH_ALL_PROXY=socks5://localhost:9999

# Access Director *thru* jumpbox (instead of being on the jumpbox)
$ bosh -e bosh-1 env
```
