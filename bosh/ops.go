package bosh

const GCPBoshDirectorEphemeralIPOps = `
- type: replace
  path: /networks/name=default/subnets/0/cloud_properties/ephemeral_external_ip?
  value: true
`

const AWSBoshDirectorEphemeralIPOps = `
- type: replace
  path: /resource_pools/name=vms/cloud_properties/auto_assign_public_ip?
  value: true
`

const AWSEncryptDiskOps = `---
- type: replace
  path: /disk_pools/name=disks/cloud_properties?
  value:
    type: gp2
    encrypted: true
    kms_key_arn: ((kms_key_arn))
`

const AzureSSHStaticIP = `
- type: replace
  path: /cloud_provider/ssh_tunnel/host
  value: ((external_ip))
`

const AzureJumpboxCpi = `
- type: replace
  path: /releases/-
  value:
    name: bosh-azure-cpi
    url: https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-azure-cpi-release?v=29
    sha1: 630901d22de58597ef8d5a23be9a5b7107d9ecb4

- type: replace
  path: /resource_pools/name=vms/stemcell?
  value:
    url: https://bosh.io/d/stemcells/bosh-azure-hyperv-ubuntu-trusty-go_agent?v=3445.11
    sha1: c70b6854ce1551fbeecfebabfcd6df5215513cad

- type: replace
  path: /resource_pools/name=vms/cloud_properties?
  value:
    instance_type: Standard_D1_v2

- type: replace
  path: /networks/name=private/subnets/0/cloud_properties?
  value:
    resource_group_name: ((resource_group_name))
    virtual_network_name: ((vnet_name))
    subnet_name: ((subnet_name))

- type: replace
  path: /networks/name=public/subnets?/-
  value:
    cloud_properties:
      resource_group_name: ((resource_group_name))

- type: replace
  path: /cloud_provider/template?
  value:
    name: azure_cpi
    release: bosh-azure-cpi

- type: replace
  path: /cloud_provider/ssh_tunnel?
  value:
    host: ((external_ip))
    port: 22
    user: vcap
    private_key: ((private_key))

- type: replace
  path: /cloud_provider/properties/azure?
  value:
    environment: AzureCloud
    subscription_id: ((subscription_id))
    tenant_id: ((tenant_id))
    client_id: ((client_id))
    client_secret: ((client_secret))
    resource_group_name: ((resource_group_name))
    storage_account_name: ((storage_account_name))
    default_security_group: ((default_security_group))
    ssh_user: vcap
    ssh_public_key: ((public_key))
`
