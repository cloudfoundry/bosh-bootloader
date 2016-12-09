package gcp

import yaml "gopkg.in/yaml.v2"

type CloudConfigInput struct {
	AZs            []string
	Tags           []string
	NetworkName    string
	SubnetworkName string
	LoadBalancer   string
}

type CloudConfigGenerator struct {
	input       CloudConfigInput
	cloudConfig CloudConfig
}

type VMExtension struct {
	Name            string                     `yaml:"name"`
	CloudProperties VMExtensionCloudProperties `yaml:"cloud_properties"`
}

type VMExtensionCloudProperties struct {
	RootDiskSizeGB int    `yaml:"root_disk_size_gb,omitempty"`
	RootDiskType   string `yaml:"root_disk_type,omitempty"`
	TargetPool     string `yaml:"target_pool,omitempty"`
}

type CloudConfig struct {
	AZs          []AZ          `yaml:"azs,omitempty"`
	Networks     []Network     `yaml:"networks,omitempty"`
	VMTypes      interface{}   `yaml:"vm_types,omitempty"`
	DiskTypes    interface{}   `yaml:"disk_types,omitempty"`
	Compilation  interface{}   `yaml:"compilation,omitempty"`
	VMExtensions []VMExtension `yaml:"vm_extensions,omitempty"`
}

var unmarshal func([]byte, interface{}) error = yaml.Unmarshal

func NewCloudConfigGenerator() CloudConfigGenerator {
	return CloudConfigGenerator{}
}

func (c CloudConfigGenerator) Generate(input CloudConfigInput) (CloudConfig, error) {
	if err := unmarshal([]byte(cloudConfigTemplate), &c.cloudConfig); err != nil {
		return CloudConfig{}, err
	}

	c.input = input

	c.generateAZs()
	if err := c.generateNetworks(); err != nil {
		return CloudConfig{}, err
	}

	if c.input.LoadBalancer != "" {
		c.cloudConfig.VMExtensions = append(c.cloudConfig.VMExtensions, VMExtension{
			Name: "lb",
			CloudProperties: VMExtensionCloudProperties{
				TargetPool: c.input.LoadBalancer,
			},
		})
	}

	return c.cloudConfig, nil
}

func (c *CloudConfigGenerator) generateAZs() {
	azsGenerator := NewAZsGenerator(c.input.AZs...)
	c.cloudConfig.AZs = azsGenerator.Generate()
}

func (c *CloudConfigGenerator) generateNetworks() error {
	azs := []string{}
	for _, az := range c.cloudConfig.AZs {
		azs = append(azs, az.Name)
	}

	networksGenerator := NewNetworksGenerator(c.input.NetworkName, c.input.SubnetworkName, c.input.Tags, azs)

	var err error
	c.cloudConfig.Networks, err = networksGenerator.Generate()
	if err != nil {
		return err
	}

	return nil
}
