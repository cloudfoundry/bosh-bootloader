
# Concourse

## Prerequisites

If one has followed the BBL steps mentioned in the [**IaaS-Specific Getting Started Guides**](https://github.com/cloudfoundry/bosh-bootloader#iaas-specific-getting-started-guides), the foundation has been created according to your IaaS of choice, which typically includes:

1. A BOSH director
2. A jumpbox
3. A set of randomly generated BOSH director credentials
4. A generated keypair allowing you to SSH into the BOSH director and any instances BOSH deploys
5. A copy of the manifest the BOSH director was deployed with
6. A basic cloud config


## Deploy Concourse Cluster

On top of these, below are the typical steps to deploy Concourse cluster.

### 1. Create Concourse LB

Let's run our scripts in the folder where one ran the `bbl up`.

```
bbl plan --lb-type concourse
bbl up
```

> Note: this will create a new IaaS Load Balancer, with some ports (80, 443, 2222, 8443, 8844) pre-configured and opened, to front Concourse web node(s).


### 2. Prepare for Concourse Deployment

```
eval "$(bbl print-env)"

export IAAS="$(cat bbl-state.json | jq -r .iaas)"
if [ "${IAAS}" = "aws" ]; then
  export EXTERNAL_HOST="$(bbl outputs | grep concourse_lb_url | cut -d ' ' -f2)"
  export STEMCELL_URL="https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-xenial-go_agent"
elif [ "${IAAS}" = "gcp" ]; then
  export EXTERNAL_HOST="$(bbl outputs | grep concourse_lb_ip | cut -d ' ' -f2)"
  export STEMCELL_URL="https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-xenial-go_agent"
else # Azure
  export EXTERNAL_HOST="$(bbl outputs | grep concourse_lb_ip | cut -d ' ' -f2)"
  export STEMCELL_URL="https://bosh.io/d/stemcells/bosh-azure-hyperv-ubuntu-xenial-go_agent"
fi

bosh upload-stemcell "${STEMCELL_URL}"
```


### 3. Customize & Deploy


```
git clone https://github.com/concourse/concourse-bosh-deployment.git

pushd concourse-bosh-deployment/cluster

    export USERNAME="username"
    export PASSWORD="super-secure-password"

    cat > ../../vars/concourse-vars-file.yml <<EOL
external_host: "${EXTERNAL_HOST}"
external_url: "https://${EXTERNAL_HOST}"
local_user:
  username: "${USERNAME}"
  password: "${PASSWORD}"
network_name: 'private'
web_instances: 1
web_network_name: 'private'
web_vm_type: 'default'
web_network_vm_extension: 'lb'
db_vm_type: 'default'
db_persistent_disk_type: '1GB'
worker_instances: 2
worker_vm_type: 'default'
worker_ephemeral_disk: '50GB_ephemeral_disk'
deployment_name: 'concourse'
EOL

  bosh deploy -d concourse concourse.yml \
    -l ../versions.yml \
    -l ../../vars/concourse-vars-file.yml \
    -o operations/basic-auth.yml \
    -o operations/privileged-http.yml \
    -o operations/privileged-https.yml \
    -o operations/tls.yml \
    -o operations/tls-vars.yml \
    -o operations/web-network-extension.yml \
    -o operations/scale.yml \
    -o operations/worker-ephemeral-disk.yml
 
popd
```


> Note: do check it out [here](https://github.com/concourse/concourse-bosh-deployment/tree/master/cluster/operations) for tons of operations files by which one can tune / customize the Concourse cluster.


### 4. Check It Out

Once it's successfully done, we can simply check it out.

View the BOSH VMs' status:

```
bosh -d concourse vms

...
Deployment 'concourse'

Instance                                     Process State  AZ  IPs       VM CID                                   VM Type  Active
db/e3921a7d-ca25-4abc-9860-8fae73625507      running        z1  10.0.1.1  vm-ebe33d11-a858-45bf-61eb-89eff5bb86f8  default  true
web/dffa0d32-6e2d-446e-838f-ecfff86f0d51     running        z1  10.0.1.0  vm-6d4585b5-fda2-4215-547a-b249de8b1384  default  true
worker/06c58730-35f1-4c2f-9bb0-10f0216f8491  running        z1  10.0.1.2  vm-026513ed-85c7-47aa-7260-0fe7c286af36  default  true
worker/3a0945f5-b59d-4ef9-8002-d7c8468c2f59  running        z1  10.0.1.3  vm-4ee75638-ba98-4474-7719-b523b3fabd23  default  true

4 vms

Succeeded
```

Open Concourse in Browser:

```
open `bosh int vars/concourse-vars-file.yml --path /external_url`
```

And login with username/password from below output:

```
bosh int vars/concourse-vars-file.yml --path /local_user
```
