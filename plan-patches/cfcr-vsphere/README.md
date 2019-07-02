# Patch: cfcr-vsphere

Steps to deploy cfcr with bbl:

1. Pick a valid IP that's within your BBL_VSPHERE_SUBNET_CIDR to be the k8s api IP.
   IPs 10 or more above the base of your cidr should be safe, but this is highly dependent on if you're going to deploy anything else to this director.

   ```bash
   export kubernetes_master_host=10.87.35.35
   ```

1. Follow the normal steps to bbl up with a patch, but provide a valid IP for your future k8s api in `cloud-config/cfcr-overrides.yml`.

    ```bash
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r bosh-bootloader/plan-patches/cfcr-vsphere/. .
    cat > cloud-config/cfcr-overrides.yml << EOF
    ---
    - type: replace
    path: /networks/name=default/subnets/0/static?
    value:
    - ${kubernetes_master_host}
    EOF
    bbl up
    eval "$(bbl print-env)"
    ```

1. export KD as your path to kubo-deployment so you can copy-paste from below if you so desire.
   be careful to check out the manifest that matches the kubo-release you downloaded above.

   ```bash
   git clone git@github.com:cloudfoundry-incubator/kubo-deployment.git
   export KD=$(pwd)/kubo-deployment
   ```

1. `bosh upload-stemcell https://bosh.io/d/stemcells/bosh-vsphere-esxi-ubuntu-xenial-go_agent?v=$(bosh int ${KD}/manifests/cfcr.yml --path=/stemcells/0/version)`

1. Update the cloud config to enable disk UUID:

   ```bash
   bosh update-config ${KD}/manifests/cloud-config/iaas/vsphere/use-vm-extensions.yml --type=cloud --name=cfcr-diskuuid
   ```

1. Deploy the cfcr manifest. Since vSphere can't provision load balancers for us, we're going to deploy with a single master with a set static IP.

   ```bash
   bosh -d cfcr deploy ${KD}/manifests/cfcr.yml \
   -o ${KD}/manifests/ops-files/iaas/vsphere/cloud-provider.yml \
   -o ${KD}/manifests/ops-files/misc/single-master.yml \
   -o ${KD}/manifests/ops-files/iaas/vsphere/master-static-ip.yml \
   -o ${KD}/manifests/ops-files/add-hostname-to-master-certificate.yml \
   -o ${KD}/manifests/ops-files/iaas/vsphere/set-working-dir-no-rp.yml \
   -o ${KD}/manifests/ops-files/iaas/vsphere/use-vm-extensions.yml \
   -v kubernetes_master_host="${kubernetes_master_host}" \
   -v api-hostname="${kubernetes_master_host}" \
   -l <(bbl outputs)
   ```

   > Note: If you'd like a multi-master cfcr, you'll need to go back to step one, select a range of 3 valid IPs, re bbl-up, and remove the `single-master.yml` opsfile from the below invocation.

1. Configure kubectl

   Then run the following to mix them together into kubectl-appropriate forms:

   ```bash
   export director_name=$(bosh int <(bbl outputs) --path=/director_name)
   export address="https://${kubernetes_master_host}:8443"
   export cluster_name="kubo:${director_name}:cfcr"
   export user_name="kubo:${director_name}:cfcr-admin"
   export context_name="kubo:${director_name}:cfcr"

   credhub login
   export admin_password=$(bosh int <(credhub get -n "${director_name}/cfcr/kubo-admin-password" --output-json) --path=/value)
   export tmp_ca_file="$(mktemp)"
   bosh int <(credhub get -n "${director_name}/cfcr/tls-kubernetes" --output-json) --path=/value/ca > "${tmp_ca_file}"

   kubectl config set-cluster "${cluster_name}" --server="${address}"  --certificate-authority="${tmp_ca_file}" --embed-certs=true
   kubectl config set-credentials "${user_name}" --token="${admin_password}"
   kubectl config set-context "${context_name}" --cluster="${cluster_name}" --user="${user_name}"
   kubectl config use-context "${context_name}"
   ```

1. Create, scale, and expose apps with the Kubernetes bootcamp docker image.

   Please note that the vSphere cloud-provider, like vSphere, does not have load balancer support built in, so you'll have to use [NodePorts](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport).

   ```bash
   kubectl run kubernetes-bootcamp --image=docker.io/jocatalin/kubernetes-bootcamp:v1 --port=8080
   kubectl get pods
   kubectl expose deployment kubernetes-bootcamp --type NodePort --name k8s-bootcamp-service
   ```

   After you've completed this, other services within your vSphere network cluster should be able to reach kubernetes-bootcamp on any worker's NodePort.
