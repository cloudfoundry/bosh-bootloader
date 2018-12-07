# Patch: cfcr-gcp

Prerequisites:

1. When running `bbl up`, ensure the service account used has the additional role 'roles/resourcemanager.projectIamAdmin'.
   This is required to create the cfcr IAM bindings for your project

Steps to deploy cfcr with bbl:

1. Supply a kubernetes master host. Your k8s api will be at this hostname.
    ```
    export kubernetes_master_host=cfcr.your-domain-here.biz
    ```
1. Follow the normal steps to bbl up with a patch
    ```
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r bosh-bootloader/plan-patches/cfcr-gcp/. .
    echo kubernetes_master_host=\"${kubernetes_master_host}\" > vars/cfcr.tfvars
    bbl up

    eval "$(bbl print-env)"
    ```

1. Upload the desired kubo-release from the github repo: https://github.com/cloudfoundry-incubator/kubo-release/releases.
   ```
   bosh upload-release <github-kubo-release-link>
   ```

1. `bosh upload-stemcell https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-xenial-go_agent`

1. Export KD as your path to kubo-deployment so you can copy-paste from below if you so desire.
   Be careful to check out the manifest that matches the kubo-release you downloaded above.
   ```
   git clone git@github.com:cloudfoundry-incubator/kubo-deployment.git
   export KD=$(pwd)/kubo-deployment
   ```

1. Deploy the cfcr manifest.
   ```
   bosh deploy -d cfcr ${KD}/manifests/cfcr.yml \
   -o ${KD}/manifests/ops-files/iaas/gcp/cloud-provider.yml \
   -v deployment_name=cfcr \
   -l <(bbl outputs)
   ```

   If you'd like to use compiled releases to speed up your deployment and worry a bit less about matching release+manifest versions check out [cfcr-compiled-deployment](https://github.com/starkandwayne/cfcr-compiled-deployment).
   ```
   export CFCRC=~/go/src/github.com/starkandwayne/cfcr-compiled-deployment
   bosh deploy -d cfcr ${CFCRC}/cfcr.yml \
   -o ${CFCRC}/ops-files/iaas/gcp/cloud-provider.yml \
   -o cfcr-ops.yml -v deployment_name=cfcr \
   -l <(bbl outputs)
   ```

1. Configure kubectl
   ```
   credhub login
   export director_name=$(bosh int <(bbl outputs) --path=/director_name)
   
   ${KD}/bin/set_kubeconfig ${director_name}/cfcr https://${kubernetes_master_host}:8443
   ```

 - Run `kubectl get nodes` to ensure kubectl was configured correctly
 - create, scale, and expose apps with the kubernetes bootcamp docker image:
   ```
   kubectl run kubernetes-bootcamp --image=docker.io/jocatalin/kubernetes-bootcamp:v1 --port=8080
   kubectl get pods
   kubectl expose deployment/kubernetes-bootcamp --type="LoadBalancer"
   kubectl get services

   export external_ip=$(kubectl get service/kubernetes-bootcamp -o jsonpath={.status.loadBalancer.ingress[0].ip})
   curl http://${external_ip}:8080
   ```

