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

const VSphereJumpboxNetworkOps = `---
- type: remove
  path: /instance_groups/name=jumpbox/networks/name=public
`

const OpenStackJumpboxKeystoneV3Ops = `---
- type: replace
  path: /releases/name=bosh-openstack-cpi/url
  value: https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-openstack-cpi-release?v=37

- type: replace
  path: /releases/name=bosh-openstack-cpi/sha1
  value: 4507b907955909bc8036afc1cf6be339b306ca03

- type: remove
  path: /cloud_provider/properties/openstack/tenant

- type: replace
  path: /cloud_provider/properties/openstack/project?
  value: ((openstack_project))

- type: replace
  path: /cloud_provider/properties/openstack/domain?
  value: ((openstack_domain))

- type: replace
  path: /cloud_provider/properties/openstack/human_readable_vm_names?
  value: true
`
