package azure

const (
	BaseOps = `
- type: replace
  path: /azs/-
  value:
    name: z1

- type: replace
  path: /azs/-
  value:
    name: z2

- type: replace
  path: /azs/-
  value:
    name: z3

- type: replace
  path: /vm_types/name=default/cloud_properties?/instance_type
  value: Standard_D1_v2

- type: replace
  path: /vm_types/name=large/cloud_properties?/instance_type
  value: Standard_D3_v2

- type: replace
  path: /disk_types/name=default/disk_size?
  value: 3000

- type: replace
  path: /disk_types/name=large/disk_size?
  value: 50_000

- type: replace
  path: /networks/-
  value:
    name: vip
    type: vip

- type: replace
  path: /compilation?
  value:
    workers: 5
    reuse_compilation_vms: true
    az: z1
    vm_type: default
    network: default
`
)
