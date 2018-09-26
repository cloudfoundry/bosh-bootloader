# Patch: cf-azure

Steps to deploy cf with bbl:

1. Follow the normal steps to bbl up with a patch.

    ```bash
    export BOSH_BOOTLOADER=<YOUR BOSH BOOTLOADER PATH>
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r $BOSH_BOOTLOADER/plan-patches/cf-azure/. .
    bbl up
    eval "$(bbl print-env)"
    ```

2. export CD as your path to cf-deployment so you can copy-paste from below if you so desire.

    ```bash
    git clone git@github.com:cloudfoundry/cf-deployment.git
    export CD=$(pwd)/cf-deployment
    ```

3. Upload the stemcell required

    ```bash
    bosh upload-stemcell https://bosh.io/d/stemcells/bosh-azure-hyperv-ubuntu-trusty-go_agent?v=3586.42
    ```

4. Upload a runtime config for bosh dns

    ```bash
    bosh update-runtime-config bosh-deployment/runtime-configs/dns.yml --name dns
    ```

5. Deploy the cf manifest.

    Notes: if you only want to do a test, use:
            -o ./ops/small-vm.yml \
          if you want to scale to one instance for each instance group, use:
            -o ./ops/scale-to-one-az.yml \
          if you want to preview the manifest before apply it. use the "bosh interpolate"
          if you want to use xip.io run the script go get it:

    ```bash
    export SYSTEM_DOMAIN="$(bosh int ./vars/director-vars-file.yml --path /cf_balancer_pub_ip).xip.io"
    ```

    ```bash
    export SYSTEM_DOMAIN="<your system domain>"
    export deployment_name="cf"
    bosh -d ${deployment_name} deploy ${CD}/cf-deployment.yml \
    --vars-store=./vars/cf-deployment-vars.yml \
    -o ${CD}/operations/use-external-blobstore.yml \
    -o ${CD}/operations/use-azure-storage-blobstore.yml \
    -o ${CD}/operations/use-compiled-releases.yml \
    -o ${CD}/operations/azure.yml \
    -o ./ops/small-vm.yml \
    -o ./ops/scale-to-one-az.yml \
    -o ./ops/rename-network-and-deployment.yml \
    -o ./ops/use-cf-resource-group.yml \
    -o ./ops/use-cf-subnet.yml \
    -v environment=AzurePublic \
    -v app_package_directory_key=cc-packages \
    -v buildpack_directory_key=cc-buildpack \
    -v droplet_directory_key=cc-droplet \
    -v resource_directory_key=cc-resource \
    -v system_domain=$SYSTEM_DOMAIN \
    -v deployment_name=${deployment_name} \
    -l <(bbl outputs)
    ```

6. Use the CF environment.

    ```bash
    export CF_ADMIN_PASSWORD=$(bosh int ./vars/cf-deployment-vars.yml --path /cf_admin_password)
    cf login -a https://api.${SYSTEM_DOMAIN} --skip-ssl-validation -u admin -p $CF_ADMIN_PASSWORD
    ```