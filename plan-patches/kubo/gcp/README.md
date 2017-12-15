To deploy kubernetes with bbl just run `bbl kubo`.

1. `bbl up` on gcp
1. add the local-dns.yml ops file to the create-director.sh script
1. `bbl up` again
1. `bosh update-runtime-config bosh_deployments/dns.yml`
1. upload releases:
- bosh-dns/0.0.11
- docker/30.1.4
- kubo/0.11.0-dev.26
- kubo-etcd/6
1. upload stemcell:
- bosh-google-kvm-ubuntu-trusty-go_agent/3468.13
1. generate the kubo manifest
- create director.yml with a combination of variables from terraform outputs, the director-vars-file.yml, and the gcp console. Example: 
```
iaas: gcp
routing_mode: iaas

# from terraform output kubo_master_lb_ip_address = 35.227.103.126
kubernetes_master_host: 35.227.103.126

# from hearsay
kubernetes_master_port: 8443

# from bbl's director-vars-file.yml
network: dr-bbl-kubo-network
project_id: cf-infra
director_name: bosh-dr-bbl-kubo

# from?
service_account_master: dr-bbl-kubo-kubo-master@cf-infra.iam.gserviceaccount.com # service account to be set on Kubo Master VMs 
service_account_worker: dr-bbl-kubo-kubo-worker@cf-infra.iam.gserviceaccount.com # service account to be set on Kubo Worker VMs 
```  
- run the kubo manifest generation script in kubo_deployment/bin/generate_kubo_manifest providing the directory with our director.yml as a flag.
1. bosh deploy the kubo manifest
1. configure kubectl
- `kubectl config set-cluster`
- `kubectl config set-credentials`
- `kubectl config set-context`
- `kubectl config use-context`
- we filled in the blanks from the kubo_deployment/bin/set_kubeconfig script:
```
address="https://${endpoint}:${port}"
admin_password=$(bosh-cli int <(credhub get -n "${director_name}/${deployment_name}/kubo-admin-password" --output-json) --path=/value)
cluster_name="kubo:${director_name}:${deployment_name}"
user_name="kubo:${director_name}:${deployment_name}-admin"
context_name="kubo:${director_name}:${deployment_name}"

kubectl config set-cluster "${cluster_name}" --server="$address" --certificate-authority="${tmp_ca_file}" --embed-certs=true
kubectl config set-credentials "${user_name}" --token="${admin_password}"
kubectl config set-context "${context_name}" --cluster="${cluster_name}" --user="${user_name}"
kubectl config use-context "${context_name}"
```
- make sure to `export https_proxy=$BOSH_ALL_PROXY` and use the master version of the credhub cli (the latest release does not respect the https_proxy environment variable)
- `kubectl get pods`
- create, scale, and expose apps with the kubernetes bootcamp docker image: docker.io/jocatalin/kubernetes-bootcamp:v1

