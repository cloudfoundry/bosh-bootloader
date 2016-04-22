package bosh

type VMExtensionsGenerators struct {
	loadBalancerExtensions []LoadBalancerExtension
}

type VMExtension struct {
	Name            string                     `yaml:"name"`
	CloudProperties VMExtensionCloudProperties `yaml:"cloud_properties"`
}

type VMExtensionCloudProperties struct {
	ELBS []string `yaml:"elbs"`
}

type LoadBalancerExtension struct {
	Name    string
	ELBName string
}

func NewVMExtensionsGenerator(loadBalancerExtensions []LoadBalancerExtension) VMExtensionsGenerators {
	return VMExtensionsGenerators{
		loadBalancerExtensions: loadBalancerExtensions,
	}
}

func (g VMExtensionsGenerators) Generate() []VMExtension {
	vmExtensions := []VMExtension{}
	for _, v := range g.loadBalancerExtensions {
		vmExtensions = append(vmExtensions, VMExtension{
			Name: v.Name,
			CloudProperties: VMExtensionCloudProperties{
				ELBS: []string{v.ELBName},
			},
		})
	}
	return vmExtensions
}
