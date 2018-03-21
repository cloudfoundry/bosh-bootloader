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
  path: /vm_extensions/name=1GB_ephemeral_disk/cloud_properties?/ephemeral_disk/size
  value: 1024

- type: replace
  path: /vm_extensions/name=5GB_ephemeral_disk/cloud_properties?/ephemeral_disk/size
  value: 5120

- type: replace
  path: /vm_extensions/name=10GB_ephemeral_disk/cloud_properties?/ephemeral_disk/size
  value: 10240

- type: replace
  path: /vm_extensions/name=50GB_ephemeral_disk/cloud_properties?/ephemeral_disk/size
  value: 51200

- type: replace
  path: /vm_extensions/name=100GB_ephemeral_disk/cloud_properties?/ephemeral_disk/size
  value: 102400

- type: replace
  path: /vm_extensions/name=500GB_ephemeral_disk/cloud_properties?/ephemeral_disk/size
  value: 512000

- type: replace
  path: /vm_extensions/name=1TB_ephemeral_disk/cloud_properties?/ephemeral_disk/size
  value: 1048576

- type: replace
  path: /vm_types/name=minimal/cloud_properties?
  value:
    ephemeral_disk:
      size: 10240
    instance_type: Standard_F1

- type: replace
  path: /vm_types/name=small/cloud_properties?
  value:
    ephemeral_disk:
      size: 10240
    instance_type: Standard_F2

- type: replace
  path: /vm_types/name=medium/cloud_properties?
  value:
    ephemeral_disk:
      size: 10240
    instance_type: Standard_F4

- type: replace
  path: /vm_types/name=large/cloud_properties?
  value:
    ephemeral_disk:
      size: 10240
    instance_type: Standard_D12_v2

- type: replace
  path: /vm_types/name=small-highmem/cloud_properties?
  value:
    ephemeral_disk:
      size: 10240
    instance_type: Standard_GS2

- type: replace
  path: /vm_types/name=sharedcpu/cloud_properties?
  value:
    ephemeral_disk:
      size: 10240
    instance_type: Standard_D1
`
)
