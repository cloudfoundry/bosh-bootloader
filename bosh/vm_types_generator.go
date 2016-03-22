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
		createVMType("m3.medium"),
		createVMType("m3.large"),
		createVMType("c3.large"),
		createVMType("c3.xlarge"),
		createVMType("c3.2xlarge"),
		createVMType("c4.large"),
		createVMType("r3.xlarge"),
		createVMType("t2.micro"),
	}
}

func createVMType(instanceType string) VMType {
	return VMType{
		Name: instanceType,
		CloudProperties: &VMTypeCloudProperties{
			InstanceType: instanceType,
			EphemeralDisk: &EphemeralDisk{
				Size: 1024,
				Type: "gp2",
			},
		},
	}
}
