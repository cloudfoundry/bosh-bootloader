package vsphere

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type OpsGenerator struct {
	terraformManager terraformManager
}

type terraformManager interface {
	GetOutputs() (terraform.Outputs, error)
}

func NewOpsGenerator(terraformManager terraformManager) OpsGenerator {
	return OpsGenerator{
		terraformManager: terraformManager,
	}
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	return `---
- type: replace
  path: /azs
  value:
  - name: z1
    cloud_properties:
      datacenters:
      - clusters: [((vcenter_cluster)): {}]
  - name: z2
    cloud_properties:
      datacenters:
      - clusters: [((vcenter_cluster)): {}]
  - name: z3
    cloud_properties:
      datacenters:
      - clusters: [((vcenter_cluster)): {}]

- type: replace
  path: /compilation
  value:
    workers: 5
    reuse_compilation_vms: true
    az: z1
    vm_type: default
    network: default

- type: replace
  path: /disk_types/name=default/disk_size?
  value: 3000

- type: replace
  path: /networks
  value:
  - name: default
    type: manual
    subnets:
    - range: ((internal_cidr))
      gateway: ((internal_gw))
      azs: [z1, z2, z3]
      dns: [8.8.8.8]
      reserved: []
      cloud_properties:
        name: ((network_name))

- type: replace
  path: /vm_types/name=default/cloud_properties?
  value:
    cpu: 2
    ram: 8_192
    disk: 30_000

- type: replace
  path: /vm_types/name=large/cloud_properties?
  value:
    cpu: 2
    ram: 8_192
    disk: 640_000

- type: replace
  path: /vm_types/name=minimal/cloud_properties?
  value:
    cpu: 1
    ram: 4096
    disk: 10240

- type: replace
  path: /vm_types/name=small/cloud_properties?
  value:
    cpu: 2
    ram: 8192
    disk: 10240

- type: replace
  path: /vm_types/name=small-highmem/cloud_properties?
  value:
    cpu: 4
    ram: 32768
    disk: 10240

- type: replace
  path: /vm_extensions/name=50GB_ephemeral_disk/cloud_properties?
  value:
    disk: 51200

- type: replace
  path: /vm_extensions/name=100GB_ephemeral_disk/cloud_properties?
  value:
    disk: 102400

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
    name: diego-ssh-proxy-network-properties
`, nil
}

func (o OpsGenerator) GenerateVars(state storage.State) (string, error) {
	terraformOutputs, err := o.terraformManager.GetOutputs()
	if err != nil {
		return "", fmt.Errorf("Get terraform outputs: %s", err)
	}
	varsBytes, err := yaml.Marshal(terraformOutputs.Map)
	if err != nil {
		panic(err) // not tested; cannot occur
	}
	return string(varsBytes), nil
}
