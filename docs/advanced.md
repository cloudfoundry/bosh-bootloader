# Advanced BOSH Configuration

### Table of Contents
* <a href='#director'>Deploy director with bosh create-env</a>
* <a href='#concourse'>Deploy concourse with bosh create-env</a>
* <a href='#opsfile'>Using an ops-file with bbl</a>


## <a name='director'></a>Deploy director with bosh create-env

**Note:** If you `bbl up --no-director`, future calls to `bbl up` will not create a director.

The process for deploying a bosh director with custom configuration on GCP is as follows:

1. Create the network and firewall rules. **Important here is the ``--no-director`` flag.**

    ```
    bbl up \
      --gcp-zone <INSERT ZONE> \
      --gcp-region <INSERT REGION> \
      --gcp-service-account-key <INSERT SERVICE ACCOUNT KEY> \
      --iaas gcp \
      --no-director
    ```

1. Use [bosh-deployment](https://github.com/cloudfoundry/bosh-deployment) to create the director.
**Important here is the ``-o external-ip-not-recommended.yml`` ops-file**
(unless you set up a tunnel to your IaaS such that you can route to the director at `10.0.0.6`).

    ```
    git clone https://github.com/cloudfoundry/bosh-deployment.git deploy
    bosh create-env deploy/bosh.yml  \
      --state ./state.json  \
      -o deploy/gcp/cpi.yml  \
      -o deploy/external-ip-not-recommended.yml \
      --vars-store ./creds.yml  \
      -l <(bbl bosh-deployment-vars)
    ```

1. Add load balancers.

    ```
    bbl create-lbs --type cf --key mykey.key --cert mycert.crt --domain cf.example.com
    ```

1. Update the cloud-config with the load balancer VM extensions.

    ```
    eval "$(bbl print-env)"
    export BOSH_CA_CERT="$(bosh int creds.yml --path /default_ca/ca)"
    export BOSH_CLIENT_SECRET="$(bosh int creds.yml --path /admin_password)"
    export BOSH_CLIENT=admin
    bosh update-cloud-config <(bbl cloud-config)
    ```

1. Deploy a manifest like [cf-deployment](https://github.com/cloudfoundry/cf-deployment).


## <a name='concourse'></a>Deploy concourse with bosh create-env

1. Create the network and firewall rules. **Important here is the `--no-director` flag.**

    ```
    bbl up \
      --gcp-zone <INSERT ZONE> \
      --gcp-region <INSERT REGION> \
      --gcp-service-account-key <INSERT SERVICE ACCOUNT KEY> \
      --iaas gcp \
      --no-director
    ```

1. Follow the deployment instructions in [concourse-deployment](https://github.com/concourse/concourse-deployment).
Use the network related variables supplied by `bbl bosh-deployment-vars`.

    ```
    git clone https://github.com/concourse/concourse-deployment.git
    bosh create-env concourse-deployment/concourse.yml  \
      --state ~/environments/concourse/state.json  \
      -o concourse-deployment/infrastructures/gcp.yml  \
      --vars-store ~/environments/concourse/creds.yml  \
      -l <(bbl --state-dir ~/environments/concourse bosh-deployment-vars | sed 's/external_ip/public_ip/')
    ```

1. Open port `4443` on the firewall rule `concourse-bosh-open`.

1. Check out your new concourse at `https://<bbl director-address>:4443`.


## <a name='opsfile'></a>Using an ops-file with bbl

#### Supply an ops-file

You can provide a single ops-file to be applied to your BOSH director wih the `--ops-file` flag in `bbl up`.

    ```
    bbl up --ops-file='/path/to/some-ops-file.yml'
    ```

The ops file will be saved in the state file for your bbl environment, so future calls to `bbl up` will continue to use the ops file.

#### Replace an ops-file

If you want to replace the ops file with another one, you can pass in a different ops file:

    ```
    bbl up --ops-file='/path/to/some-other-ops-file.yml'
    ```

#### Remove an ops-file

If you want to remove the ops file, you can supply the flag again with an empty YAML file or an empty string:

    ```
    bbl up --ops-file=''
    ```
