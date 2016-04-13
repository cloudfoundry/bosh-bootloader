package bosh

type VMExtensionsGenerators struct {
	lbType string
}

type VMExtension struct {
	Name            string                     `yaml:"name"`
	CloudProperties VMExtensionCloudProperties `yaml:"cloud_properties"`
}

type VMExtensionCloudProperties struct {
	ELBS []string `yaml:"elbs"`
}

func NewVMExtensionsGenerator(lbType string) VMExtensionsGenerators {
	return VMExtensionsGenerators{
		lbType: lbType,
	}
}

func (g VMExtensionsGenerators) Generate() []VMExtension {
	return []VMExtension{
		{
			Name: "lb",
			CloudProperties: VMExtensionCloudProperties{
				ELBS: []string{g.lbType},
			},
		},
	}
}
