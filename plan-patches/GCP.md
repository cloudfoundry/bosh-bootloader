# GCP Plan Patches

Plan patches can be used to customize the IAAS
environment and bosh director that is created by
`bbl up`.

A patch is a directory with a set of files
organized in the same hierarchy as the bbl-state dir.

* <a href='#bosh-lite-gcp'>Deploy BOSH-Lite</a>
* <a href='#restricted-instance-groups-gcp'>Create 2 Instance Groups</a>
* <a href='#iso-segs-gcp'>Add Isolation Segments</a>
* <a href='#tf-backend-gcp'>Use GCS Bucket for Terraform State</a>
* <a href='#byobastion-gcp'>Bring Your Own Bastion</a>
* <a href='#cfcr-gcp'>Deploy CFCR</a>

## <a name='bosh-lite-gcp'></a>bosh-lite-gcp

To create a bosh lite on gcp, the files in `bosh-lite-gcp`
should be copied to your bbl state directory.

The steps might look like such:

```
mkdir banana-env && cd banana-env

bbl plan --name banana-env

cp -r bosh-bootloader/plan-patches/bosh-lite-gcp/. .

bbl up
```


## <a name='restricted-instance-groups-gcp'></a>restricted-instance-groups-gcp

To create two instance groups instead of an instance group for every zone on gcp,
you can use the steps above with the `restricted-instance-groups-gcp` patch
provided here.

## <a name='iso-segs-gcp'></a>iso-segs-gcp

Creates a single routing isolation segment on GCP, including dedicated load balancers and firewall rules.

```
cp -r bosh-bootloader/plan-patches/iso-segs-gcp/. some-env/
bbl up
```

Disclaimer: this is a testing/development quality patch.  It has not been subject to a security review -- the firewall rules may not be fully locked down.
Please don't run it in production!


## <a name='tf-backend-gcp'></a>tf-backend-gcp

Stores the terraform state in a bucket in Google Cloud Storage.

```
cp -r bosh-bootloader/plan-patches/tf-backend-gcp/. .
```

Since the backend configuration is loaded by Terraform extremely early (before
the core of Teraform can be initialized), there can be no interplations in the backend
configuration template. Instead of providing vars for the bucket to a `gcs_backend_override.tfvars`,
the values for the bucket name, credential path, and state prefix must be provided directly
in the backend configuration template.

Modify `terraform/gcs_backend_override.tf` to provide the name of the bucket, the path to
the GCP service account key, and a prefix for the environment inside the bucket.

Then you can bbl up.

```
bbl up
```

## <a name='byobastion-gcp'></a>byobastion-gcp

To use your own bastion on gcp, the files in `byobastion-gcp`
should be copied to your bbl state directory.

Why would you do this?

You want to use a vm in your vpc that can serve
as the bastion to a bbl'd up director, but that makes
using a bbl'd up jumpbox feel redundant. This patch
will deploy a director that can be accessed directly from the
bastion without a jumpbox, but not from the larger internets.

The steps might look like such:

1. In the gcp console, create a network and a vm, put it in a 10.0.0.0/16 subnet.
1. ssh to that vm, install bbl, terraform, and the bosh cli (and its dependencies.)
1. Create your bbl-state and apply this plan patch
    ```
    mkdir banana-env && cd banana-env

    cp -r bosh-bootloader/plan-patches/byobastion-gcp/. .

    bbl plan --name banana-env
    ```

1. Configure the patch and bbl up:
    ```
    vim vars/bastion.tfvars # fill out variables with the network, subnet, and external ip name you've made in the gcp console

    bbl up # will exit 1
    ```
    Note: `bbl up` fails while trying to update the cloud-config
    because it assumes there is a jumpbox and tries to contact the director
    via an ssh tunnel + proxy through it. To upload a cloud config without proxying
    through your (nonexistant) jumpbox, you can run:

1. Upload your cloud config
    ```
    eval "$(bbl print-env | grep -v BOSH_ALL_PROXY)"

    bosh update-cloud-config cloud-config/cloud-config.yml -o cloud-config/ops.yml
    ```
1. Once your director is deployed, you can target it with:
    ```
    eval "$(bbl print-env | grep -v BOSH_ALL_PROXY)"

    bosh deploy ...
    ```

## <a name='cfcr-gcp'></a>cfcr-gcp

Steps to deploy cfcr with bbl:

1. Follow the normal steps to bbl up with a patch
    ```
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r bosh-bootloader/plan-patches/cfcr-gcp/. .
    bbl up
    ```

1. `bosh upload-release https://storage.googleapis.com/kubo-public/kubo-release-latest.tgz`

1. `bosh upload-stemcell https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-trusty-go_agent?v=3468.21`

1. Create cfcr-vars.yml: with a combination of variables from terraform outputs, the director-vars-file.yml, and the gcp console. Example:

    ```
    bosh int cfcr-vars-template.yml \
    --vars-file vars/director-vars-file.yml \
    --vars-file vars/cloud-config-vars.yml \
    -v  project_id="${BBL_GCP_PROJECT_ID}" \
    > cfcr-vars.yml
    ```
    this file will contain the additional variables necessary for a cfcr deployment.

1. Deploy the cfcr manifest

   ```
   bosh deploy -d cfcr ~/kubo-deployment/manifests/cfcr.yml \
   -o ~/kubo-deployment/manifests/ops-files/iaas/gcp/cloud-provider.yml \
   -o ./kubo-ops.yml \
   --vars-file cfcr-vars.yml
   ```

1. Configure kubectl

   Note: at the the time of PRing this patch, credhub+boshcli+bbl support requires some work make sure to:
     1. set the JUMPBOX_PUBLIC_IP environment variable to the jumpbox public ip found in BOSH_ALL_PROXY
     1. open an ssh tunnel to your jumpbox `ssh -f -N -o StrictHostKeyChecking=no -o ServerAliveInterval=300 -D 61943 jumpbox@${JUMPBOX_PUBLIC_IP} -i /var/folders/sy/ypl_k9gd29bc4b8g3w6hw0840000gn/T/bosh-jumpbox415804785/bosh_jumpbox_private.key`
     1. `export https_proxy=socks5://localhost:61943`
     1. Use the latest version of the credhub cli (old versions do not respect the https_proxy environment variable)
     1. `unset https_proxy` when you are done running credhub commands because it interferes with the bosh cli

   Export `deployment_name`, `director_name`, `kubernetes_master_host`, and `kubernetes_master_port` from your`cfcar-vars.yml` file.
   Then run the following to mix them together into kubectl-appropriate forms:
   ```
   # right now, we don't support TLS verification of the kubernetes master, so we also don't need to run these commmands.
   # export tmp_ca_file="$(mktemp)"
   # bosh int <(credhub get -n "${director_name}/${deployment_name}/tls-kubernetes" --output-json) --path=/value/ca > "${tmp_ca_file}"

   export address="https://${kubernetes_master_host}:${kubernetes_master_port}"
   export admin_password=$(bosh int <(credhub get -n "${director_name}/${deployment_name}/kubo-admin-password" --output-json) --path=/value)
   export cluster_name="kubo:${director_name}:${deployment_name}"
   export user_name="kubo:${director_name}:${deployment_name}-admin"
   export context_name="kubo:${director_name}:${deployment_name}"
   ```
   ```
   kubectl config set-cluster "${cluster_name}" --server="${address}" --insecure-skip-tls-verify=true
   kubectl config set-credentials "${user_name}" --token="${admin_password}"
   kubectl config set-context "${context_name}" --cluster="${cluster_name}" --user="${user_name}"
   kubectl config use-context "${context_name}"
   ```

 - `kubectl get pods`
 - create, scale, and expose apps with the kubernetes bootcamp docker image:
   ```
   kubectl run kubernetes-bootcamp --image=docker.io/jocatalin/kubernetes-bootcamp:v1 --port=8080
   kubectl get pods
   kubectl expose deployment/kubernetes-bootcamp --type="LoadBalancer"
   kubectl get services
   # get EXTERNAL-IP for kubernetes-bootcamp
   curl http://${EXTERNAL-IP}:8080
   ```

