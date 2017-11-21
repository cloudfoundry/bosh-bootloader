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
  path: /disk_types
  value:
  - name: default
    disk_size: 3000
  - name: large
    disk_size: 50_000

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
  path: /vm_types
  value:
  - name: default
    cloud_properties:
      cpu: 2
      ram: 8_192
      disk: 30_000
  - name: large
    cloud_properties:
      cpu: 2
      ram: 8_192
      disk: 640_000

- type: remove
  path: /vm_extensions
`, nil
}

type VarsYAML struct {
	InternalCIDR   string `yaml:"internal_cidr,omitempty"`
	InternalGW     string `yaml:"internal_gw,omitempty"`
	NetworkName    string `yaml:"network_name,omitempty"`
	VCenterCluster string `yaml:"vcenter_cluster,omitempty"`
}

func (o OpsGenerator) GenerateVars(state storage.State) (string, error) {
	terraformOutputs, err := o.terraformManager.GetOutputs()
	if err != nil {
		return "", fmt.Errorf("Get terraform outputs: %s", err)
	}
	varsYAML := VarsYAML{
		InternalCIDR:   terraformOutputs.GetString("internal_cidr"),
		InternalGW:     terraformOutputs.GetString("internal_gw"),
		NetworkName:    terraformOutputs.GetString("network_name"),
		VCenterCluster: terraformOutputs.GetString("vcenter_cluster"),
	}
	varsBytes, err := yaml.Marshal(varsYAML)
	if err != nil {
		panic(err) // not tested; cannot occur
	}
	return string(varsBytes), nil
}
