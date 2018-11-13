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

const VSphereJumpboxNetworkOps = `---
- type: remove
  path: /instance_groups/name=jumpbox/networks/name=public
`
