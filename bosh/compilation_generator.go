package bosh

type CompilationGenerator struct{}

type Compilation struct {
	Workers             int    `yaml:"workers,omitempty"`
	Network             string `yaml:"network,omitempty"`
	AZ                  string `yaml:"az,omitempty"`
	ReuseCompilationVMs bool   `yaml:"reuse_compilation_vms,omitempty"`
	VMType              string `yaml:"vm_type,omitempty"`
}

func NewCompilationGenerator() CompilationGenerator {
	return CompilationGenerator{}
}

func (CompilationGenerator) Generate() *Compilation {
	return &Compilation{
		Workers:             3,
		Network:             "concourse",
		AZ:                  "z1",
		ReuseCompilationVMs: true,
		VMType:              "default",
	}
}
