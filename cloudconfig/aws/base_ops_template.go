package aws

const (
	BaseOps = `
- type: replace
  path: /compilation/vm_type
  value: c3.large

- type: replace
  path: /disk_types/name=1GB/cloud_properties?
  value:
    type: gp2
    encrypted: true

- type: replace
  path: /disk_types/name=5GB/cloud_properties?
  value:
    type: gp2
    encrypted: true

- type: replace
  path: /disk_types/name=10GB/cloud_properties?
  value:
    type: gp2
    encrypted: true

- type: replace
  path: /disk_types/name=50GB/cloud_properties?
  value:
    type: gp2
    encrypted: true

- type: replace
  path: /disk_types/name=100GB/cloud_properties?
  value:
    type: gp2
    encrypted: true

- type: replace
  path: /disk_types/name=500GB/cloud_properties?
  value:
    type: gp2
    encrypted: true

- type: replace
  path: /disk_types/name=1TB/cloud_properties?
  value:
    type: gp2
    encrypted: true

- type: replace
  path: /vm_types/name=default/cloud_properties?
  value:
    instance_type: m3.medium
    ephemeral_disk:
      size: 10240
      type: gp2

- type: replace
  path: /vm_types/name=sharedcpu/cloud_properties?
  value:
    instance_type: t2.small
    ephemeral_disk:
      size: 10240
      type: gp2

- type: replace
  path: /vm_types/name=small/cloud_properties?
  value:
    instance_type: m3.large
    ephemeral_disk:
      size: 10240
      type: gp2

- type: replace
  path: /vm_types/name=medium/cloud_properties?
  value:
    instance_type: m4.xlarge
    ephemeral_disk:
      size: 10240
      type: gp2

- type: replace
  path: /vm_types/name=large/cloud_properties?
  value:
    instance_type: m4.2xlarge
    ephemeral_disk:
      size: 10240
      type: gp2

- type: replace
  path: /vm_types/name=extra-large/cloud_properties?
  value:
    instance_type: m4.4xlarge
    ephemeral_disk:
      size: 10240
      type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m3.medium
    cloud_properties:
      instance_type: m3.medium
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m3.large
    cloud_properties:
      instance_type: m3.large
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m3.xlarge
    cloud_properties:
      instance_type: m3.xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m3.2xlarge
    cloud_properties:
      instance_type: m3.2xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m4.large
    cloud_properties:
      instance_type: m4.large
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m4.xlarge
    cloud_properties:
      instance_type: m4.xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m4.2xlarge
    cloud_properties:
      instance_type: m4.2xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m4.4xlarge
    cloud_properties:
      instance_type: m4.4xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: m4.10xlarge
    cloud_properties:
      instance_type: m4.10xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c3.large
    cloud_properties:
      instance_type: c3.large
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c3.xlarge
    cloud_properties:
      instance_type: c3.xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c3.2xlarge
    cloud_properties:
      instance_type: c3.2xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c3.4xlarge
    cloud_properties:
      instance_type: c3.4xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c3.8xlarge
    cloud_properties:
      instance_type: c3.8xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c4.large
    cloud_properties:
      instance_type: c4.large
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c4.xlarge
    cloud_properties:
      instance_type: c4.xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c4.2xlarge
    cloud_properties:
      instance_type: c4.2xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c4.4xlarge
    cloud_properties:
      instance_type: c4.4xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: c4.8xlarge
    cloud_properties:
      instance_type: c4.8xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: r3.large
    cloud_properties:
      instance_type: r3.large
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: r3.xlarge
    cloud_properties:
      instance_type: r3.xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: r3.2xlarge
    cloud_properties:
      instance_type: r3.2xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: r3.4xlarge
    cloud_properties:
      instance_type: r3.4xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: r3.8xlarge
    cloud_properties:
      instance_type: r3.8xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: t2.nano
    cloud_properties:
      instance_type: t2.nano
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: t2.micro
    cloud_properties:
      instance_type: t2.micro
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: t2.small
    cloud_properties:
      instance_type: t2.small
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: t2.medium
    cloud_properties:
      instance_type: t2.medium
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: t2.large
    cloud_properties:
      instance_type: t2.large
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: small-highmem
    cloud_properties:
      instance_type: r3.xlarge
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_types/-
  value:
    name: small-highcpu
    cloud_properties:
      instance_type: c3.large
      ephemeral_disk:
        size: 10240
        type: gp2

- type: replace
  path: /vm_extensions/name=1GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk:
      size: 1024
      type: gp2

- type: replace
  path: /vm_extensions/name=5GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk:
      size: 5120
      type: gp2

- type: replace
  path: /vm_extensions/name=10GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk:
      size: 10240
      type: gp2

- type: replace
  path: /vm_extensions/name=50GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk:
      size: 51200
      type: gp2

- type: replace
  path: /vm_extensions/name=100GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk:
      size: 102400
      type: gp2

- type: replace
  path: /vm_extensions/name=500GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk:
      size: 512000
      type: gp2

- type: replace
  path: /vm_extensions/name=1TB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk:
      size: 1048576
      type: gp2
`
)
