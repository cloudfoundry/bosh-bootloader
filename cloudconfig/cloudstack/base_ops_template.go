package cloudstack

const (
	BaseOps = `
- type: replace
  path: /azs
  value:
  - name: z1
    cloud_properties: {}
  - name: z2
    cloud_properties: {}
  - name: z3
    cloud_properties: {}

- type: replace
  path: /compilation
  value:
    workers: 5
    reuse_compilation_vms: true
    vm_type: default
    az: z1
    network: compilation

- type: replace
  path: /disk_types/name=default?
  value:
    name: default
    cloud_properties:
      disk_offering: shared.custom
    disk_size: 1024

- type: replace
  path: /disk_types/name=1GB?
  value:
    name: 1GB
    cloud_properties:
      disk_offering: shared.custom
    disk_size: 1024

- type: replace
  path: /disk_types/name=5GB?
  value:
    name: 5GB
    cloud_properties:
      disk_offering: shared.custom
    disk_size: 5120

- type: replace
  path: /disk_types/name=10GB?
  value:
    name: 10GB
    cloud_properties:
      disk_offering: shared.small
    disk_size: 10240

- type: replace
  path: /disk_types/name=50GB?
  value:
    name: 50GB
    cloud_properties:
      disk_offering: shared.medium
    disk_size: 51200

- type: replace
  path: /disk_types/name=100GB?
  value:
    name: 100GB
    cloud_properties:
      disk_offering: shared.large
    disk_size: 102400

- type: replace
  path: /disk_types/name=500GB?
  value:
    name: 500GB
    cloud_properties:
      disk_offering: shared.xxlarge
    disk_size: 512000

- type: replace
  path: /disk_types/name=1TB?
  value:
    name: 1TB
    cloud_properties:
      disk_offering: shared.custom
    disk_size: 1048576

- type: replace
  path: /vm_types/name=default/cloud_properties?
  value:
    compute_offering: shared.medium

- type: replace
  path: /vm_types/name=minimal/cloud_properties?
  value:
    compute_offering: shared.small

- type: replace
  path: /vm_types/name=small/cloud_properties?
  value:
    compute_offering: shared.medium

- type: replace
  path: /vm_types/name=small-highmem/cloud_properties?
  value:
    compute_offering: shared.xlarge

- type: replace
  path: /vm_types/name=medium/cloud_properties?
  value:
    compute_offering: shared.xmedium

- type: replace
  path: /vm_types/name=large/cloud_properties?
  value:
    compute_offering: shared.xlarge

- type: replace
  path: /vm_types/name=extra-large/cloud_properties?
  value:
    compute_offering: shared.xlarge

- type: replace
  path: /vm_extensions/name=1GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk_offering: shared.custom
    disk: 1024

- type: replace
  path: /vm_extensions/name=5GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk_offering: shared.custom
    disk: 5120

- type: replace
  path: /vm_extensions/name=10GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk_offering: shared.small

- type: replace
  path: /vm_extensions/name=50GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk_offering: shared.medium

- type: replace
  path: /vm_extensions/name=100GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk_offering: shared.large

- type: replace
  path: /vm_extensions/name=500GB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk_offering: shared.xxlarge

- type: replace
  path: /vm_extensions/name=1TB_ephemeral_disk/cloud_properties?
  value:
    ephemeral_disk_offering: shared.custom
    disk: 1048576

- type: replace
  path: /vm_extensions/-
  value:
    name: cf-router-network-properties

- type: replace
  path: /vm_extensions/-
  value:
    name: cf-tcp-router-network-properties

- type: replace
  path: /vm_extensions/-
  value:
    name: diego-ssh-proxy-network-properties`
)
