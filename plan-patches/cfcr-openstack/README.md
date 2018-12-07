# Patch: cfcr-openstack

## Known Issues

- This patch does not support dns.
- This patch is tied to `kubo-release v0.16.0` and will change with the next release
- `kubo-deployment` is using keystone v2 and `bosh-deployment` is using keystone v3.
We have to change the auth url (that we expect to be v3) provided to bbl to use the v2 endpoint.

## Steps

Steps to deploy cfcr with bbl:

1. Pick a valid floating IP that's available within your openstack account and export it so we can attach it to the k8s master.
   ```
   export kubernetes_master_host=some-ip
   ```

1. Find and export the project ID guid associated with $BBL_OPENSTACK_PROJECT
   ```
   export OPENSTACK_RPOJECT_ID=some-uid
   ```

1. Follow the normal steps to bbl up with a patch.
    ```
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r bosh-bootloader/plan-patches/cfcr-openstack/. .
    bbl up
    eval "$(bbl print-env)"
    ```

1. Export `KD` as your path to `kubo-deployment` so you can copy-paste from below if you so desire.
   be careful to check out the manifest that matches the kubo-release you uploaded above.
   ```
   git clone git@github.com:cloudfoundry-incubator/kubo-deployment.git
   export KD=$(pwd)/kubo-deployment
   ```

1. `bosh upload-stemcell https://bosh.io/stemcells/bosh-openstack-esxi-ubuntu-trusty-go_agent?v=$(bosh int ${KD}/manifests/cfcr.yml --path=/stemcells/0/version)`

1. Deploy the cfcr manifest. Since openstack can't provision load balancers for
us, we're going to deploy with a single master with a set static IP.

   ```
   bosh -d cfcr deploy ${KD}/manifests/cfcr.yml \
   -o ${KD}/manifests/ops-files/iaas/openstack/cloud-provider.yml \
   -o ${KD}/manifests/ops-files/iaas/openstack/master-static-ip.yml \
   -v kubernetes_master_host=${kubernetes_master_host} \
   -v openstack_username=${BBL_OPENSTACK_USERNAME} \
   -v openstack_password=${BBL_OPENSTACK_PASSWORD} \
   -v openstack_project_id=${OPENSTACK_PROJECT_ID} \
   -l <(bbl outputs) \
   -v auth_url=$(sed 's|v3|v2.0|' <(echo $BBL_OPENSTACK_AUTH_URL))
   ```

   > Note If you'd like a multi-master cfcr, you'll need to go back to step one,
   > select a range of 3 valid IPs, re bbl-up, and remove the `single-master.yml` opsfile from the below invokation.
   > The master-static-ip.yml opsfile in kubo-deployment might not play well with 3 static IPs.

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
   ```

   If you want to have a tls-secured kubernetes api, you'll need to add credhub's generated CA to your trusted CAs. We'll leave that as an excercise for the operator.
   ```
   # export tmp_ca_file="$(mktemp)"
   # bosh int <(credhub get -n "${director_name}/cfcr/tls-kubernetes" --output-json) --path=/value/ca > "${tmp_ca_file}"
   ```

   ```
   kubectl config set-cluster "${cluster_name}" --server="${address}" --insecure-skip-tls-verify=true
   kubectl config set-credentials "${user_name}" --token="${admin_password}"
   kubectl config set-context "${context_name}" --cluster="${cluster_name}" --user="${user_name}"
   kubectl config use-context "${context_name}"
   ```

1. Create, scale, and expose apps with the kubernetes bootcamp docker image.
Please note that the openstack cloud-provider, like openstack, does not have load balancer support built in, so you'll have to use nodeports.

   ```
   kubectl run kubernetes-bootcamp --image=docker.io/jocatalin/kubernetes-bootcamp:v1 --port=8080
   kubectl get pods
   kubectl expose deployment kubernetes-bootcamp --type NodePort --name k8s-bootcamp-service
   ```

   After you've completed this, other services within your openstack cluster should be able to reach kubernetes-bootcamp on any worker's NodePort.
