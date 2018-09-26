# Patch: cf-azure

Steps to deploy cf-lite with bbl:

1. Follow the normal steps to bbl up with a patch.

    ```bash
    export BOSH_BOOTLOADER=<YOUR BOSH BOOTLOADER PATH>
    mkdir banana-env && cd banana-env
    bbl plan --name banana-env
    cp -r $BOSH_BOOTLOADER/plan-patches/cf-lite-azure/. .
    bbl up
    eval "$(bbl print-env)"
    ```

2. export CD as your path to cf-lite-deployment so you can copy-paste from below if you so desire.

    ```bash
    git clone git@github.com:virtualcloudfoundry/cf-lite-deployment.git
    export CD=$(pwd)/cf-lite-deployment
    ```

3. Upload the stemcell required

    ```bash
    bosh upload-stemcell https://bosh.io/d/stemcells/bosh-azure-hyperv-ubuntu-xenial-go_agent?v=97.17
    ```

4. Upload the all-in-one-release to handle the dnat rules needed.

    ```bash
    git clone git@github.com:virtualcloudfoundry/cf-lite-release.git
    pushd ./cf-lite-release
    bosh create-release && bosh upload-release
    popd
    ```

5. Deploy the cf manifest.

    Notes: if you only want to do a test, use:
          -o ./ops/small-vm.yml \
        if you want to preview the manifest before apply it. use the "bosh interpolate"
        if you want to use xip.io run the script go get it:

    ```bash
    export SYSTEM_DOMAIN="$(bosh int ./vars/director-vars-file.yml --path /cf_balancer_pub_ip).xip.io"
    ```

    ```bash
    export SYSTEM_DOMAIN="<your system domain>"
    bosh -n -d cf-lite deploy ${CD}/cf-lite-deployment.yml \
    --vars-store=./vars/cf-lite-deployment-vars.yml \
    -o ${CD}/operations/use-external-blobstore.yml \
    -o ${CD}/operations/use-azure-storage-blobstore.yml \
    -o ${CD}/operations/azure.yml \
    -o ./ops/use-cf-resource-group.yml \
    -o ./ops/use-cf-subnet.yml \
    -v environment=AzurePublic \
    -v app_package_directory_key=cc-packages \
    -v buildpack_directory_key=cc-buildpack \
    -v droplet_directory_key=cc-droplet \
    -v resource_directory_key=cc-resource \
    -v system_domain=$SYSTEM_DOMAIN \
    -l <(bbl outputs)
    ```

6. Use the CF environment.

    ```bash
    export CF_ADMIN_PASSWORD=$(bosh int ./vars/cf-lite-deployment-vars.yml --path /cf_admin_password)
    cf login -a https://api.${SYSTEM_DOMAIN} --skip-ssl-validation -u admin -p $CF_ADMIN_PASSWORD
    ```