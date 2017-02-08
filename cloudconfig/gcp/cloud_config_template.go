package gcp

const cloudConfigTemplate = `
compilation:
  az: z1
  network: private
  reuse_compilation_vms: true
  vm_type: n1-highcpu-8
  workers: 6
  vm_extensions:
  - 100GB_ephemeral_disk

disk_types:
- name: 1GB
  disk_size: 1024
  cloud_properties:
    type: pd-ssd
    encrypted: true
- name: 5GB
  disk_size: 5120
  cloud_properties:
    type: pd-ssd
    encrypted: true
- name: 10GB
  disk_size: 10240
  cloud_properties:
    type: pd-ssd
    encrypted: true
- name: 50GB
  disk_size: 51200
  cloud_properties:
    type: pd-ssd
    encrypted: true
- name: 100GB
  disk_size: 102400
  cloud_properties:
    type: pd-ssd
    encrypted: true
- name: 500GB
  disk_size: 512000
  cloud_properties:
    type: pd-ssd
    encrypted: true
- name: 1TB
  disk_size: 1048576
  cloud_properties:
    type: pd-ssd
    encrypted: true

vm_types:
- name: default
  cloud_properties:
    machine_type: n1-standard-1
    root_disk_size_gb: 10
    root_disk_type: pd-ssd

- name: n1-standard-1
  cloud_properties:
    machine_type: n1-standard-1
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-standard-2
  cloud_properties:
    machine_type: n1-standard-2
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-standard-4
  cloud_properties:
    machine_type: n1-standard-4
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-standard-8
  cloud_properties:
    machine_type: n1-standard-8
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-standard-16
  cloud_properties:
    machine_type: n1-standard-16
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-standard-32
  cloud_properties:
    machine_type: n1-standard-32
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highmem-2
  cloud_properties:
    machine_type: n1-highmem-2
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highmem-4
  cloud_properties:
    machine_type: n1-highmem-4
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highmem-8
  cloud_properties:
    machine_type: n1-highmem-8
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highmem-16
  cloud_properties:
    machine_type: n1-highmem-16
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highmem-32
  cloud_properties:
    machine_type: n1-highmem-32
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highcpu-2
  cloud_properties:
    machine_type: n1-highcpu-2
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highcpu-4
  cloud_properties:
    machine_type: n1-highcpu-4
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highcpu-8
  cloud_properties:
    machine_type: n1-highcpu-8
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highcpu-16
  cloud_properties:
    machine_type: n1-highcpu-16
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: n1-highcpu-32
  cloud_properties:
    machine_type: n1-highcpu-32
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: f1-micro
  cloud_properties:
    machine_type: f1-micro
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: g1-small
  cloud_properties:
    machine_type: g1-small
    root_disk_size_gb: 10
    root_disk_type: pd-ssd

- name: m3.medium
  cloud_properties:
    machine_type: n1-standard-1
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: m3.large
  cloud_properties:
    machine_type: n1-standard-2
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: c3.large
  cloud_properties:
    machine_type: n1-highcpu-2
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: r3.xlarge
  cloud_properties:
    machine_type: n1-highmem-4
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: t2.small
  cloud_properties:
    machine_type: g1-small
    root_disk_size_gb: 10
    root_disk_type: pd-ssd

vm_extensions:
- name: 5GB_ephemeral_disk
  cloud_properties:
    root_disk_size_gb: 5
    root_disk_type: pd-ssd
- name: 10GB_ephemeral_disk
  cloud_properties:
    root_disk_size_gb: 10
    root_disk_type: pd-ssd
- name: 50GB_ephemeral_disk
  cloud_properties:
    root_disk_size_gb: 50
    root_disk_type: pd-ssd
- name: 100GB_ephemeral_disk
  cloud_properties:
    root_disk_size_gb: 100
    root_disk_type: pd-ssd
- name: 500GB_ephemeral_disk
  cloud_properties:
    root_disk_size_gb: 500
    root_disk_type: pd-ssd
- name: 1TB_ephemeral_disk
  cloud_properties:
    root_disk_size_gb: 1000
    root_disk_type: pd-ssd
- name: internet-required
  cloud_properties:
    ephemeral_external_ip: true
- name: internet-not-required
  cloud_properties:
    ephemeral_external_ip: false
`
