package bosh

type CloudConfigInput struct {
	AZs     []string
	Subnets []SubnetInput
	LBs     []LoadBalancerExtension
}

type SubnetInput struct {
	AZ             string
	Subnet         string
	CIDR           string
	SecurityGroups []string
}

type CloudConfigGenerator struct {
	input       CloudConfigInput
	cloudConfig CloudConfig
}

func NewCloudConfigGenerator() CloudConfigGenerator {
	return CloudConfigGenerator{}
}

func (c CloudConfigGenerator) Generate(input CloudConfigInput) (CloudConfig, error) {
	c.input = input

	c.generateAZs()
	c.generateVMTypes()
	c.generateDiskTypes()
	c.generateCompilation()
	err := c.generateNetworks()
	if err != nil {
		return CloudConfig{}, err
	}

	if len(c.input.LBs) > 0 {
		c.generateVMExtensions()
	}

	return c.cloudConfig, nil
}

func (c *CloudConfigGenerator) generateVMExtensions() {
	c.cloudConfig.VMExtensions = NewVMExtensionsGenerator(c.input.LBs).Generate()
}

func (c *CloudConfigGenerator) generateAZs() {
	azsGenerator := NewAZsGenerator(c.input.AZs...)
	c.cloudConfig.AZs = azsGenerator.Generate()
}

func (c *CloudConfigGenerator) generateVMTypes() {
	vmTypesGenerator := NewVMTypesGenerator()
	c.cloudConfig.VMTypes = vmTypesGenerator.Generate()
}

func (c *CloudConfigGenerator) generateDiskTypes() {
	diskTypesGenerator := NewDiskTypesGenerator()
	c.cloudConfig.DiskTypes = diskTypesGenerator.Generate()
}

func (c *CloudConfigGenerator) generateCompilation() {
	compilationGenerator := NewCompilationGenerator()
	c.cloudConfig.Compilation = compilationGenerator.Generate()
}

func (c *CloudConfigGenerator) generateNetworks() error {
	azAssociations := map[string]string{}
	for _, az := range c.cloudConfig.AZs {
		azAssociations[az.CloudProperties.AvailabilityZone] = az.Name
	}

	networksGenerator := NewNetworksGenerator(c.input.Subnets, azAssociations)
	var err error
	c.cloudConfig.Networks, err = networksGenerator.Generate()
	if err != nil {
		return err
	}
	return nil
}
