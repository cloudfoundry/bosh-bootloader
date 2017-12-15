Steps to deploy kubo with bbl:

1. `bbl plan` on gcp
1. add the local-dns.yml ops file to the create-director.sh script
1. move the `kubo_override.tf` file from `plan-patches/kubo/gcp` to the `terraform` directory of the bbl state dir.  
   Note: `dns-addresses.yml` and `bosh-admin-client.yml` are used in the kubo scripts, but we did not use them.
1. `bbl up`
1. move the `kubo-ops.yml` file from `plan-patches/kubo/gcp` to the `cloud-config` directory of the bbl state dir.
1. edit the `kubo-ops.yml` file to replace these variables with the correct values from the terraform output:
    - `service_account_master`
    - `service_account_worker`
    - `target_pool`
1. `bbl up` again to update the cloud-config
1. `bosh update-runtime-config bosh_deployments/runtime-configs/dns.yml`
1. `bosh upload-release https://storage.googleapis.com/kubo-public/kubo-deployment-latest.tgz`
1. `bosh upload-stemcell https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-trusty-go_agent?v=3468.13`
1. generate the kubo manifest
    - create director.yml with a combination of variables from terraform outputs, the director-vars-file.yml, and the gcp console. Example:  
    ```
    iaas: gcp
    routing_mode: iaas
    
    # from terraform outputs
    kubernetes_master_host: 35.227.103.126
    service_account_master: dr-bbl-kubo-kubo-master@cf-infra.iam.gserviceaccount.com
    service_account_worker: dr-bbl-kubo-kubo-worker@cf-infra.iam.gserviceaccount.com
    
    # from bbl's director-vars-file.yml
    network: dr-bbl-kubo-network
    project_id: cf-infra
    director_name: bosh-dr-bbl-kubo

    # from hearsay
    kubernetes_master_port: 8443
    ```
    - run the kubo manifest generation script:  
      `{KUBO_DEPLOYMENT_PATH}/bin/generate_kubo_manifest [KUBO_ENV] [DEPLOYMENT_NAME] [DIRECTOR_UUID]`
      - KUBO_ENV: the directory with our director.yml
      - DEPLOYMENT_NAME: some name of your choosing, eg: my-kubo
      - DIRECTOR_UUID: can be found with `bosh env`
1. bosh deploy the kubo manifest
1. configure kubectl
    - `kubectl config set-cluster`
    - `kubectl config set-credentials`
    - `kubectl config set-context`
    - `kubectl config use-context`
    - we followed the kubo_deployment/bin/set_kubeconfig script:  
      Note: make sure to `export https_proxy=$BOSH_ALL_PROXY` and use the master version of the credhub cli (the latest release does not respect the https_proxy environment variable)
      ```
      tmp_ca_file="$(mktemp)"
      bosh-cli int <(credhub get -n "${director_name}/${deployment_name}/tls-kubernetes" --output-json) --path=/value/ca > "${tmp_ca_file}"

      address="https://${kubernetes_master_host}:${kubernetes_master_port}"
      admin_password=$(bosh-cli int <(credhub get -n "${director_name}/${deployment_name}/kubo-admin-password" --output-json) --path=/value)
      cluster_name="kubo:${director_name}:${deployment_name}"
      user_name="kubo:${director_name}:${deployment_name}-admin"
      context_name="kubo:${director_name}:${deployment_name}"
      
      kubectl config set-cluster "${cluster_name}" --server="$address" --certificate-authority="${tmp_ca_file}" --embed-certs=true
      kubectl config set-credentials "${user_name}" --token="${admin_password}"
      kubectl config set-context "${context_name}" --cluster="${cluster_name}" --user="${user_name}"
      kubectl config use-context "${context_name}"
      ```
    - `kubectl get pods`
    - create, scale, and expose apps with the kubernetes bootcamp docker image: `docker.io/jocatalin/kubernetes-bootcamp:v1`

