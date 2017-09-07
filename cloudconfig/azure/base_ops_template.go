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
  path: /disk_types/name=default/disk_size
  value: 3000
`
)
