package bosh

type VMExtensionsGenerators struct {
	loadBalancerExtensions []LoadBalancerExtension
}

type VMExtension struct {
	Name            string                     `yaml:"name"`
	CloudProperties VMExtensionCloudProperties `yaml:"cloud_properties"`
}

type VMExtensionCloudProperties struct {
	ELBS           []string                  `yaml:"elbs,omitempty"`
	SecurityGroups []string                  `yaml:"security_groups,omitempty"`
	EphemeralDisk  *VMExtensionEphemeralDisk `yaml:"ephemeral_disk,omitempty"`
}

type VMExtensionEphemeralDisk struct {
	Size int    `yaml:"size"`
	Type string `yaml:"type"`
}

type LoadBalancerExtension struct {
	Name           string
	ELBName        string
	SecurityGroups []string
}

func NewVMExtensionsGenerator(loadBalancerExtensions []LoadBalancerExtension) VMExtensionsGenerators {
	return VMExtensionsGenerators{
		loadBalancerExtensions: loadBalancerExtensions,
	}
}

func (g VMExtensionsGenerators) Generate() []VMExtension {
	vmExtensions := []VMExtension{
		{
			Name: "5GB_ephemeral_disk",
			CloudProperties: VMExtensionCloudProperties{
				EphemeralDisk: &VMExtensionEphemeralDisk{
					Size: 5120,
					Type: "gp2",
				},
			},
		},
		{
			Name: "10GB_ephemeral_disk",
			CloudProperties: VMExtensionCloudProperties{
				EphemeralDisk: &VMExtensionEphemeralDisk{
					Size: 10240,
					Type: "gp2",
				},
			},
		},
		{
			Name: "50GB_ephemeral_disk",
			CloudProperties: VMExtensionCloudProperties{
				EphemeralDisk: &VMExtensionEphemeralDisk{
					Size: 51200,
					Type: "gp2",
				},
			},
		},
		{
			Name: "100GB_ephemeral_disk",
			CloudProperties: VMExtensionCloudProperties{
				EphemeralDisk: &VMExtensionEphemeralDisk{
					Size: 102400,
					Type: "gp2",
				},
			},
		},
		{
			Name: "500GB_ephemeral_disk",
			CloudProperties: VMExtensionCloudProperties{
				EphemeralDisk: &VMExtensionEphemeralDisk{
					Size: 512000,
					Type: "gp2",
				},
			},
		},
		{
			Name: "1TB_ephemeral_disk",
			CloudProperties: VMExtensionCloudProperties{
				EphemeralDisk: &VMExtensionEphemeralDisk{
					Size: 1048576,
					Type: "gp2",
				},
			},
		},
	}

	for _, v := range g.loadBalancerExtensions {
		vmExtensions = append(vmExtensions, VMExtension{
			Name: v.Name,
			CloudProperties: VMExtensionCloudProperties{
				ELBS:           []string{v.ELBName},
				SecurityGroups: v.SecurityGroups,
			},
		})
	}

	return vmExtensions
}
