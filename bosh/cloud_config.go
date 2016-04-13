package bosh

type CloudConfig struct {
	AZs          []AZ          `yaml:"azs,omitempty"`
	VMTypes      []VMType      `yaml:"vm_types,omitempty"`
	DiskTypes    []DiskType    `yaml:"disk_types,omitempty"`
	Compilation  *Compilation  `yaml:"compilation,omitempty"`
	Networks     []Network     `yaml:"networks,omitempty"`
	VMExtensions []VMExtension `yaml:"vm_extensions,omitempty"`
}
