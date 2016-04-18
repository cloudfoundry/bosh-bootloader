package bosh

type VMExtensionsGenerators struct {
	lbTypes map[string]string
}

type VMExtension struct {
	Name            string                     `yaml:"name"`
	CloudProperties VMExtensionCloudProperties `yaml:"cloud_properties"`
}

type VMExtensionCloudProperties struct {
	ELBS []string `yaml:"elbs"`
}

func NewVMExtensionsGenerator(lbTypes map[string]string) VMExtensionsGenerators {
	return VMExtensionsGenerators{
		lbTypes: lbTypes,
	}
}

func (g VMExtensionsGenerators) Generate() []VMExtension {
	vmExtensions := []VMExtension{}
	for k, v := range g.lbTypes {
		vmExtensions = append(vmExtensions, VMExtension{
			Name: k,
			CloudProperties: VMExtensionCloudProperties{
				ELBS: []string{v},
			},
		})
	}
	return vmExtensions
}
