Steps to deploy kubo with bbl:

1. Follow the normal steps to bbl up with a patch
    ```
    mkdir some-env && cd some-env
    bbl plan --name some-env
    cp -r /path/to/this-patch-dir/. .
    bbl up
    ```
1. `bosh update-runtime-config bosh_deployments/runtime-configs/dns.yml`
1. `bosh upload-release https://storage.googleapis.com/kubo-public/kubo-deployment-latest.tgz`
1. `bosh upload-stemcell https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-trusty-go_agent?v=3468.13`
1. generate the kubo manifest
    - create director.yml: with a combination of variables from terraform outputs, the director-vars-file.yml, and the gcp console. Example:  
    ```
    bosh int director-template.yml --vars-file vars/director-vars-file.yml --vars-file vars/cloud-config-vars.yml > director.yml
    ```
    - run the kubo manifest generation script:  
      `{KUBO_DEPLOYMENT_PATH}/bin/generate_kubo_manifest [KUBO_ENV] kubo [DIRECTOR_UUID] > kubo-deployment.yml`
      - KUBO_ENV: the directory with our director.yml
      - DIRECTOR_UUID: can be found with `bosh env`
1. bosh deploy the kubo manifest
   ```
   bosh deploy -d kubo kubo-deployment.yml
   ```
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

      deployment_name=kubo
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
    - create, scale, and expose apps with the kubernetes bootcamp docker image:
      ```
      kubectl run kubernetes-bootcamp --image=docker.io/jocatalin/kubernetes-bootcamp:v1 --port=8080
      kubectl get pods
      kubectl expose deployment/kubernetes-bootcamp --type="LoadBalancer"
      kubectl get services
      # get EXTERNAL-IP for kubernetes-bootcamp
      curl http://${EXTERNAL-IP}:8080
      ```

