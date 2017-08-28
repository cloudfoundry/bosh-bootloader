# Advanced BOSH Configuration

For many users, especially cloudfoundry core teams, you may want to specify configuration options for the bosh director that bosh-bootloader does not expose:
* External Credential Store
* CA cert for the director
* ARP cache flushing

In previous versions of bosh-bootloader the only option was to allow bosh-bootloader to manage your bosh director entirely,
but as of v3.0 we have introduced the option to only have bbl pave the IaaS but leave director management up to you.

You can also deploy things directly like [concourse](#deploying-concourse-deployment-directly-with-bosh-create-env) yourself without a bosh director.

Alternatively, you can now [supply an ops file](#using-an-ops-file-with-bbl) to `bbl up` for a customized BOSH director that is still managed by `bbl`.

## Using bbl and bosh-deployment to deploy your own BOSH director

The process to create a director with custom configuration is as follows:

1. ``bbl up`` with the ``--no-director`` flag
2. Use ``bbl bosh-deployment-vars`` and [bosh-deployment](https://github.com/cloudfoundry/bosh-deployment) to create a director
3. ``bbl create-lbs`` to get load balancers
4. Use ``bbl cloud-config`` and the bosh cli to upload a cloud config containing VM extensions for your load balancer(s)

Note: If you `bbl up --no-director`, future calls to `bbl up` will not create director.

### A concrete example, with full arguments supplied

First we create our network and firewall rules. Important here is the ``--no-director`` flag.
```
bbl up --gcp-zone us-west1-a --gcp-region us-west1 --gcp-service-account-key service-account.key.json --gcp-project-id my-project-14478532 --iaas gcp --no-director
```


Next we use bosh-deployment to create the director. Take special care that ``-o external-ip-not-recommended.yml`` is supplied (or set up a tunnel to your IaaS such that you can route to 10.0.0.6, the director).
```
git clone https://github.com/cloudfoundry/bosh-deployment.git deploy
bosh create-env deploy/bosh.yml  \
  --state ./state.json  \
  -o deploy/gcp/cpi.yml  \
  -o deploy/external-ip-not-recommended.yml \
  --vars-store ./creds.yml  \
  -l <(bbl bosh-deployment-vars)
```

Now add load balancers
```
bbl create-lbs --type cf --key mykey.key --cert mycert.crt --domain cf.example.com
```

Then upload the load balancer VM extensions to your cloud-config
```
eval "$(bbl print-env)"
export BOSH_CA_CERT="$(bosh int creds.yml --path /default_ca/ca)"
export BOSH_CLIENT_SECRET="$(bosh int creds.yml --path /admin_password)"
export BOSH_CLIENT=admin
bosh update-cloud-config <(bbl cloud-config)
```

Finally deploy a bosh deployment manifest like [cf-deployment](https://github.com/cloudfoundry/cf-deployment)

### AWS Example

First create AWS infrastructure but do not create `BOSH Director`

```
$ bbl up \
	--aws-access-key-id <INSERT ACCESS KEY ID> \
	--aws-secret-access-key <INSERT SECRET ACCESS KEY> \
	--aws-region eu-central-1 \
	--iaas aws \
	--no-director
```

Now clone manifest, make necessary modifications and deploy BOSH.

```
$ git clone https://github.com/cloudfoundry/bosh-deployment.git deploy
$ bosh create-env deploy/bosh.yml  \
  --state ./state.json  \
  -o deploy/aws/cpi.yml  \
  -o deploy/external-ip-with-registry-not-recommended.yml \
  --vars-store ./creds.yml  \
  -l <(bbl bosh-deployment-vars)
```

To verify list available deployments (should be empty):

```
$ eval "$(bbl print-env)"
$ export BOSH_CA_CERT="$(bosh int creds.yml --path /default_ca/ca)"
$ export BOSH_CLIENT_SECRET="$(bosh int creds.yml --path /admin_password)"
$ export BOSH_CLIENT=admin
$ bosh deployments
```

## Deploying concourse-deployment directly with BOSH create-env

The ``--no-director`` flag can also be used to create the necessary IaaS configuration for [concourse-deployment](https://github.com/concourse/concourse-deployment), a minimal version of concourse deployed with `bosh create-env`.
```
mkdir -p ~/environments/concourse/
bbl up --state-dir ~/my-bbl-states/concourse/ \
       --gcp-zone us-west1-a --gcp-region us-west1 \
       --gcp-service-account-key service-account.key.json
       --gcp-project-id my-project-14478532 \
       --iaas gcp --name concourse \
       --no-director
```

Next we follow the deployment instructions in [concourse-deployment](https://github.com/concourse/concourse-deployment), however, many of the network related variables are supplied by `bosh-deployment-vars`
```
git clone https://github.com/concourse/concourse-deployment.git 
bosh create-env concourse-deployment/concourse.yml  \
  --state ~/environments/concourse/state.json  \
  -o concourse-deployment/infrastructures/gcp.yml  \
  --vars-store ~/environments/concourse/creds.yml  \
  -l <(bbl --state-dir ~/environments/concourse bosh-deployment-vars | sed 's/external_ip/public_ip/')
```

Last, open port 4443 on the firewall rule concourse-bosh-open. Now you should be able to see your new concourse at `https://<bbl director-address>:4443`.

## Using an ops file with bbl

An ops file for your BOSH director can be supplied with the `--ops-file` flag to the `bbl up` command.

```bash
bbl up --ops-file='/path/to/some-ops-file.yml'
```

The ops file will be saved in the state file for your bbl environment, so future calls to `bbl up` will continue to use the ops file.

If you want to replace the ops file with another one, you can pass in a different ops file:

```bash
bbl up --ops-file='/path/to/some-other-ops-file.yml'
```

If you want to remove the ops file, you can supply the flag again with an empty YAML file or an empty string:

```bash
bbl up --ops-file=''
```
