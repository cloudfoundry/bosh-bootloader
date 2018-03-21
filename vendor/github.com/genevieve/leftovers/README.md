# Leftovers

Go cli/library for cleaning up **orphaned IAAS resources**.

It will **prompt you before deleting** any resource, ie:

```
$ leftovers --filter reindeer

Are you sure you want to delete firewall bbl-env-reindeer? (y/N)
```



## Why you might be here?
- You `terraform apply`'d way back when and lost your `terraform.tfstate`
- You used the console or cli to create some infrastructure and want to clean up
- Your acceptance tests in CI failed, the container disappeared, and
infrastructure resources were tragically orphaned



## Installation

[Install go.](https://golang.org/doc/install) Then:

```
$  go get -u github.com/genevieve/leftovers/cmd/leftovers
```

**OR**

```
brew tap genevieve/tap
brew install leftovers
```



## Usage

```
Usage:
  leftovers [OPTIONS]

Application Options:
  -i, --iaas=                     The IAAS for clean up. (default: aws) [$BBL_IAAS]
  -n, --no-confirm                Destroy resources without prompting. This is dangerous, make good choices!
  -f, --filter=                   Filtering resources by an environment name.
      --aws-access-key-id=        AWS access key id. [$BBL_AWS_ACCESS_KEY_ID]
      --aws-secret-access-key=    AWS secret access key. [$BBL_AWS_SECRET_ACCESS_KEY]
      --aws-region=               AWS region. [$BBL_AWS_REGION]
      --azure-client-id=          Azure client id. [$BBL_AZURE_CLIENT_ID]
      --azure-client-secret=      Azure client secret. [$BBL_AZURE_CLIENT_SECRET]
      --azure-tenant-id=          Azure tenant id. [$BBL_AZURE_TENANT_ID]
      --azure-subscription-id=    Azure subscription id. [$BBL_AZURE_SUBSCRIPTION_ID]
      --gcp-service-account-key=  GCP service account key path. [$BBL_GCP_SERVICE_ACCOUNT_KEY]
      --vsphere-vcenter-ip=       vSphere vCenter IP address. [$BBL_VSPHERE_VCENTER_IP]
      --vsphere-vcenter-password= vSphere vCenter password. [$BBL_VSPHERE_VCENTER_PASSWORD]
      --vsphere-vcenter-user=     vSphere vCenter username. [$BBL_VSPHERE_VCENTER_USER]
      --vsphere-vcenter-dc=       vSphere vCenter datacenter. [$BBL_VSPHERE_VCENTER_DC]

Help Options:
  -h, --help                     Show this help message
```



## What's being deleted by IAAS:

### AWS
#### What can you delete with this?

  ```diff
  + iam instance profiles (& detaching roles)
  + iam roles
  + iam role policies
  + iam user policies
  + iam server certificates
  + ec2 volumes
  + ec2 tags
  + ec2 key pairs
  + ec2 instances
  + ec2 security groups
  + ec2 vpcs
  + ec2 subnets
  + ec2 route tables
  + ec2 internet gateways
  + ec2 network interfaces
  + elb load balancers
  + elbv2 load balancers
  + elbv2 target groups
  + s3 buckets
  + rds db instances
  + rds db subnet groups
  ```

#### What's up next?

  ```diff
  - iam group policies
  - ec2 eips
  ```

### Microsoft Azure

#### What can you delete with this?

  ```diff
  + resource groups
  ```

### GCP

#### What can you delete with this?

  ```diff
  + compute addresses
  + compute global addresses
  + compute backend services
  + compute disks
  + compute firewalls
  + compute forwarding rules
  + compute global forwarding rules
  + compute global health checks
  + compute http health checks
  + compute https health checks
  + compute images
  + compute subnetworks
  + compute networks
  + compute target pools
  + compute target https proxies
  + compute target http proxies
  + compute url maps
  + compute vm instances
  + compute vm instance groups
  + dns managed zones
  + dns record sets
  ```
#### What's up next?

  ```diff
  - compute routes
  - compute vm instance templates
  - compute snapshots
  ```

### vSphere

#### What can you delete with this?

  ```diff
  + virtual machines
  + empty folders
  ```

#### What's up next?

