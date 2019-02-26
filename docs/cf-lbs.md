# Cloud Foundry Load Balancers

## AWS
`bbl` creates 3 load balancers on AWS.

1. **cf-ssh-lb**

    * In the cloud-config, this lb is referenced with the vm extension `diego-ssh-proxy-network-properties`.
    * In cf-deployment, this vm extension will be associated with the `scheduler` vm.
    * It forwards `TCP:2222` to `TCP:2222`.

1. **cf-tcp-lb**

    * In the cloud-config, this lb is referenced with the vm extension `cf-tcp-router-network-properties`.
    * In cf-deployment, this vm extension will be associated with the `tcp-router` vm.
    * It forwards `TCP:1024-1123` to `TCP:1024-1123`.

1. **cf-router-lb**

    * In the cloud-config, this lb is referenced with the vm extension `cf-router-network-properties`.
    * In cf-deployment, this vm extension will be associated with the `router` vm.
    * It forwards:
        - `HTTP:80`   to `HTTP:80`
        - `HTTPS:443` to `HTTP:80`

## GCP
`bbl` creates 4 load balancers on GCP.

1. **cf-router-lb**

    * In the cloud-config, this lb is referenced with the vm extension `cf-router-network-properties`.
    * In cf-deployment, this vm extension will be associated with the `router` vm.
    * Configuration:
        - Compute address
        - Backend service with an instance group per availability zone
        - Instance group per availability zone allowing `https:443`
        - Firewall rule allowing `tcp:80` & `tcp:443` to the backend service
        - Health check for `tcp:8080` & `tcp:80`

1. **cf-ws-lb**
    * In the cloud-config, this lb is referenced with the vm extension `cf-router-network-properties`.
    * In cf-deployment, this vm extension will be associated with the `router` vm.
    * Configuration:
        - Compute address
        - Target pool
        - Forwarding rule allowing `tcp:443` to the target pool
        - Forwarding rule allowing `tcp:80` to the target pool

1. **cf-tcp-router-lb**
    * In the cloud-config, this lb is referenced with the vm extension `cf-tcp-router-network-properties`.
    * In cf-deployment, this vm extension will be associated with the `tcp-router` vm.
    * Configuration:
        - Compute address
        - Target pool
        - Firewall rule allowing `tcp:1024-32768` to the target pool
        - Forwarding rule for `tcp:1024-32768` to the target pool

1. **cf-ssh-proxy-lb**
    * In the cloud-config, this lb is referenced with the vm extension `diego-ssh-proxy-network-properties`.
    * In cf-deployment, this vm extension will be associated with the `scheduler` vm.
    * Configuration:
        - Compute address
        - Target pool
        - Firewall rule allowing `tcp:2222` to the target pool
        - Forwarding rule for `tcp:2222` to the target pool

## Microsoft Azure
`bbl` creates an application gateway on Microsoft Azure.

1. **cf-app-gateway**
    * In the cloud-config, this lb is referenced with the vm extension `cf-router-network-properties`.
    * In cf-deployment, this vm extension will be associated with the `router` vm.
    * Configuration:
        - Public IP
        - Application Gateway
        - Network Security Rules
        - Network Security Group

## vSphere
N/A.

## OpenStack
N/A.
