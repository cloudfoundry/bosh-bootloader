# Patch: cfcr-vsphere

Steps to deploy cfcr with bbl:

1. Pick a valid IP that's within your BBL_VSPHERE_SUBNET to be the k8s api IP.
   IPs 10 or more above the base of your cidr should be safe, but this is highly dependent on if you're going to deploy anything else to this director.
   ```
   export kubernetes_master_host=10.87.35.35
   ```
1. Follow the normal steps to bbl up with a patch, but provide a valid IP for your future k8s api in `cloud-config/cfcr-overrides.yml`.
    ```
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r bosh-bootloader/plan-patches/cfcr-vsphere/. .
    cat > cloud-config/cfcr-overrides.yml << EOF
---
- type: replace
  path: /networks/name=default/subnets/0/static?
  value:
  - $kubernetes_master_host
EOF
    bbl up
    eval "$(bbl print-env)"
    ```

1. `bosh upload-release https://storage.googleapis.com/kubo-public/kubo-release-latest.tgz`

1. `bosh upload-stemcell https://bosh.io/stemcells/bosh-vsphere-esxi-ubuntu-trusty-go_agent`

1. export KD as your path to kubo-deployment so you can copy-paste from below if you so desire.
   be careful to check out the manifest that matches the kubo-release you downloaded above.
   ```
   git clone git@github.com:cloudfoundry-incubator/kubo-deployment.git
   export KD=$(pwd)/kubo-deployment
   ```

1. Deploy the cfcr manifest. Since vsphere can't provision load balancers for us, we're going to deploy with a single master with a set static IP.
   ```
   bosh -d cfcr deploy ${KD}/manifests/cfcr.yml \
   -o ${KD}/manifests/ops-files/iaas/vsphere/cloud-provider.yml \
   -o ${KD}/manifests/ops-files/misc/single-master.yml \
   -o ${KD}/manifests/ops-files/iaas/vsphere/master-static-ip.yml \
   -v kubernetes_master_host="${kubernetes_master_host}" \
   -o ${KD}/manifests/ops-files/iaas/vsphere/set-working-dir-no-rp.yml \
   -l <(bbl outputs)
   ```
   > Note If you'd like a multi-master cfcr, you'll need to go back to step one, select a range of 3 valid IPs, re bbl-up, and remove the `single-master.yml` opsfile from the below invokation.

1. Configure kubectl

   Then run the following to mix them together into kubectl-appropriate forms:
   ```
   export director_name=$(bosh int <(bbl outputs) --path=/director_name)
   export address="https://${kubernetes_master_host}:8443"
   export cluster_name="kubo:${director_name}:cfcr"
   export user_name="kubo:${director_name}:cfcr-admin"
   export context_name="kubo:${director_name}:cfcr"

   credhub login
   export admin_password=$(bosh int <(credhub get -n "${director_name}/cfcr/kubo-admin-password" --output-json) --path=/value)

   # right now, we don't support TLS verification of the kubernetes master, so we also don't need to run these commmands.
   # export tmp_ca_file="$(mktemp)"
   # bosh int <(credhub get -n "${director_name}/cfcr/tls-kubernetes" --output-json) --path=/value/ca > "${tmp_ca_file}"
   ```
   ```
   kubectl config set-cluster "${cluster_name}" --server="${address}" --insecure-skip-tls-verify=true
   kubectl config set-credentials "${user_name}" --token="${admin_password}"
   kubectl config set-context "${context_name}" --cluster="${cluster_name}" --user="${user_name}"
   kubectl config use-context "${context_name}"
   ```

1. create, scale, and expose apps with the kubernetes bootcamp docker image. please note that the vsphere cloud-provider, like vsphere, does not have load balancer support built in, so you'll have to use nodeports.
   ```
   kubectl run kubernetes-bootcamp --image=docker.io/jocatalin/kubernetes-bootcamp:v1 --port=8080
   kubectl get pods
   kubectl expose deployment kubernetes-bootcamp --type NodePort --name k8s-bootcamp-service
   ```
   to curl this app, you'll want to find its nodeport with `kubectl get service k8s-bootcamp-service -o wide`.
   and its IP by correlating the node listed in `kubectl get pods -o wide` against the external IP from `kubectl get nodes -o wide`.

