package gcp

const (
	BaseOps = `
- type: replace
  path: /compilation/vm_type
  value: e2-highcpu-4

- type: replace
  path: /disk_types/name=default/cloud_properties?
  value:
    type: pd-balanced
    encrypted: true

- type: replace
  path: /disk_types/name=1GB/cloud_properties?
  value:
    type: pd-balanced
    encrypted: true

- type: replace
  path: /disk_types/name=5GB/cloud_properties?
  value:
    type: pd-balanced
    encrypted: true

- type: replace
  path: /disk_types/name=10GB/cloud_properties?
  value:
    type: pd-balanced
    encrypted: true

- type: replace
  path: /disk_types/name=50GB/cloud_properties?
  value:
    type: pd-balanced
    encrypted: true

- type: replace
  path: /disk_types/name=100GB/cloud_properties?
  value:
    type: pd-balanced
    encrypted: true

- type: replace
  path: /disk_types/name=500GB/cloud_properties?
  value:
    type: pd-balanced
    encrypted: true

- type: replace
  path: /disk_types/name=1TB/cloud_properties?
  value:
    type: pd-balanced
    encrypted: true

- type: replace
  path: /vm_types/name=default/cloud_properties?
  value:
    machine_type: n1-standard-1
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/name=minimal/cloud_properties?
  value:
    machine_type: e2-small
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/name=sharedcpu/cloud_properties?
  value:
    machine_type: g1-small
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/name=small/cloud_properties?
  value:
    machine_type: e2-medium
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/name=small-highmem/cloud_properties?
  value:
    machine_type: e2-highmem-2
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/name=small-highcpu?/cloud_properties
  value:
    machine_type: n1-highcpu-2
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/name=medium/cloud_properties?
  value:
    machine_type: n1-standard-4
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/name=large/cloud_properties?
  value:
    machine_type: n1-standard-8
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/name=extra-large/cloud_properties?
  value:
    machine_type: n1-standard-16
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-standard-1
    cloud_properties:
      machine_type: n1-standard-1
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-standard-2
    cloud_properties:
      machine_type: n1-standard-2
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-standard-4
    cloud_properties:
      machine_type: n1-standard-4
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-standard-8
    cloud_properties:
      machine_type: n1-standard-8
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-standard-16
    cloud_properties:
      machine_type: n1-standard-16
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-standard-32
    cloud_properties:
      machine_type: n1-standard-32
      root_disk_size_gb: 10
      root_disk_type: pd-balanced


- type: replace
  path: /vm_types/-
  value:
    name: n1-highmem-2
    cloud_properties:
      machine_type: n1-highmem-2
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highmem-4
    cloud_properties:
      machine_type: n1-highmem-4
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highmem-8
    cloud_properties:
      machine_type: n1-highmem-8
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highmem-16
    cloud_properties:
      machine_type: n1-highmem-16
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highmem-32
    cloud_properties:
      machine_type: n1-highmem-32
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highcpu-2
    cloud_properties:
      machine_type: n1-highcpu-2
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highcpu-4
    cloud_properties:
      machine_type: n1-highcpu-4
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highcpu-8
    cloud_properties:
      machine_type: n1-highcpu-8
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highcpu-16
    cloud_properties:
      machine_type: n1-highcpu-16
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: n1-highcpu-32
    cloud_properties:
      machine_type: n1-highcpu-32
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: f1-micro
    cloud_properties:
      machine_type: f1-micro
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: g1-small
    cloud_properties:
      machine_type: g1-small
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: m3.medium
    cloud_properties:
      machine_type: n1-standard-1
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: m3.large
    cloud_properties:
      machine_type: n1-standard-2
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: c3.large
    cloud_properties:
      machine_type: n1-highcpu-2
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: r3.xlarge
    cloud_properties:
      machine_type: n1-highmem-4
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_types/-
  value:
    name: t2.small
    cloud_properties:
      machine_type: g1-small
      root_disk_size_gb: 10
      root_disk_type: pd-balanced

- type: replace
  path: /vm_extensions/name=1GB_ephemeral_disk/cloud_properties?
  value:
    root_disk_size_gb: 1
    root_disk_type: pd-balanced

- type: replace
  path: /vm_extensions/name=5GB_ephemeral_disk/cloud_properties?
  value:
    root_disk_size_gb: 5
    root_disk_type: pd-balanced

- type: replace
  path: /vm_extensions/name=10GB_ephemeral_disk/cloud_properties?
  value:
    root_disk_size_gb: 10
    root_disk_type: pd-balanced

- type: replace
  path: /vm_extensions/name=50GB_ephemeral_disk/cloud_properties?
  value:
    root_disk_size_gb: 50
    root_disk_type: pd-balanced

- type: replace
  path: /vm_extensions/name=100GB_ephemeral_disk/cloud_properties?
  value:
    root_disk_size_gb: 100
    root_disk_type: pd-balanced

- type: replace
  path: /vm_extensions/name=500GB_ephemeral_disk/cloud_properties?
  value:
    root_disk_size_gb: 500
    root_disk_type: pd-balanced

- type: replace
  path: /vm_extensions/name=1TB_ephemeral_disk/cloud_properties?
  value:
    root_disk_size_gb: 1000
    root_disk_type: pd-balanced

- type: replace
  path: /vm_extensions/-
  value:
    name: internet-required
    cloud_properties:
      ephemeral_external_ip: true

- type: replace
  path: /vm_extensions/-
  value:
    name: internet-not-required
    cloud_properties:
      ephemeral_external_ip: false

- type: replace
  path: /vm_extensions/-
  value:
    name: preemptible
    cloud_properties:
      preemptible: true
`
)
