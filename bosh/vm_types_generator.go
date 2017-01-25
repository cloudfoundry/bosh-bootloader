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
		createVMTypeWithCustomName("default", "m3.medium"),
		createVMType("m3.medium"),
		createVMType("m3.large"),
		createVMType("m3.xlarge"),
		createVMType("m3.2xlarge"),
		createVMType("m4.large"),
		createVMType("m4.xlarge"),
		createVMType("m4.2xlarge"),
		createVMType("m4.4xlarge"),
		createVMType("m4.10xlarge"),
		createVMType("c3.large"),
		createVMType("c3.xlarge"),
		createVMType("c3.2xlarge"),
		createVMType("c3.4xlarge"),
		createVMType("c3.8xlarge"),
		createVMType("c4.large"),
		createVMType("c4.xlarge"),
		createVMType("c4.2xlarge"),
		createVMType("c4.4xlarge"),
		createVMType("c4.8xlarge"),
		createVMType("r3.large"),
		createVMType("r3.xlarge"),
		createVMType("r3.2xlarge"),
		createVMType("r3.4xlarge"),
		createVMType("r3.8xlarge"),
		createVMType("t2.nano"),
		createVMType("t2.micro"),
		createVMType("t2.small"),
		createVMType("t2.medium"),
		createVMType("t2.large"),
	}
}

func createVMType(instanceType string) VMType {
	return createVMTypeWithCustomName(instanceType, instanceType)
}

func createVMTypeWithCustomName(name, instanceType string) VMType {
	return VMType{
		Name: name,
		CloudProperties: &VMTypeCloudProperties{
			InstanceType: instanceType,
			EphemeralDisk: &EphemeralDisk{
				Size: 10240,
				Type: "gp2",
			},
		},
	}
}
