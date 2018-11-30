# Patch: cfcr-aws

Steps to deploy CFCR with bbl:

1. Follow the normal steps to bbl up with a patch
    ```bash
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r bosh-bootloader/plan-patches/cfcr-aws/. .

    bbl up
    eval "$(bbl print-env)"
    ```

1. export KD as your path to kubo-deployment so you can copy-paste from below if you so desire
   ```bash
   git clone git@github.com:cloudfoundry-incubator/kubo-deployment.git
   export KD=$(pwd)/kubo-deployment
   ```

1. Upload the stemcell.
   ```bash
   bosh upload-stemcell https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-xenial-go_agent?v=$(bosh int ${KD}/manifests/cfcr.yml --path=/stemcells/0/version)
   ```

1. Apply custom cloud config
   ```bash
   bosh update-config --name cfcr \
   ${KD}/manifests/cloud-config/iaas/aws/use-vm-extensions.yml \
   --var deployment_name=cfcr \
   --type cloud \
   --vars-file <(bbl outputs)
   ```

1. Deploy the CFCR manifest
   ```bash
   bosh deploy -d cfcr ${KD}/manifests/cfcr.yml \
   -o ${KD}/manifests/ops-files/iaas/aws/cloud-provider.yml \
   -o ${KD}/manifests/ops-files/use-vm-extensions.yml \
   -o ${KD}/manifests/ops-files/add-hostname-to-master-certificate.yml \
   -o ${KD}/manifests/ops-files/iaas/aws/lb.yml \
   --var deployment_name=cfcr \
   -l <(bbl outputs)
   ```

1. Configure kubectl

   Then run the following to mix them together into kubectl-appropriate forms:
   ```bash
   credhub login
   export director_name=$(bosh int <(bbl outputs) --path=/director_name)

   ${KD}/bin/set_kubeconfig ${director_name}/cfcr https://$(bosh int <(bbl outputs) --path /api-hostname):8443
   ```

 - Run `kubectl get pods` to check kubectl was configured correctly.
 - create, scale, and expose apps with the kubernetes bootcamp docker image:
   ```bash
   kubectl run kubernetes-bootcamp --image=gcr.io/google-samples/kubernetes-bootcamp:v1 --port=8080
   kubectl get pods
   kubectl expose deployment/kubernetes-bootcamp --type="LoadBalancer"
   kubectl get services
   # get LOAD_BALANCER_DNS_NAME for kubernetes-bootcamp
   curl http://${LOAD_BALANCER_DNS_NAME}
   ```

