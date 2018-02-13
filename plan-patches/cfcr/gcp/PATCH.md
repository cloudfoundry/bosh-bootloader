# Patch: cfcr-gcp

Steps to deploy cfcr with bbl:

1. Follow the normal steps to bbl up with a patch
    ```
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r bosh-bootloader/plan-patches/cfcr/gcp/. .
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

