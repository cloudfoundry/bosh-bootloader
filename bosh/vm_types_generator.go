package bosh

type VMTypesGenerator struct {
}

type VMType struct {
	Name            string                 `yaml:"name,omitempty"`
	CloudProperties *VMTypeCloudProperties `yaml:"cloud_properties,omitempty"`
}

type VMTypeCloudProperties struct {
	InstanceType  string         `yaml:"instance_type,omitempty"`
	EphemeralDisk *EphemeralDisk `yaml:"ephemeral_disk,omitempty"`
}

type EphemeralDisk struct {
	Size int    `yaml:"size,omitempty"`
	Type string `yaml:"type,omitempty"`
}

func NewVMTypesGenerator() VMTypesGenerator {
	return VMTypesGenerator{}
}

func (g VMTypesGenerator) Generate() []VMType {
	return []VMType{
		{
			Name: "default",
			CloudProperties: &VMTypeCloudProperties{
				InstanceType: "m3.medium",
				EphemeralDisk: &EphemeralDisk{
					Size: 1024,
					Type: "gp2",
				},
			},
		},
	}
}
