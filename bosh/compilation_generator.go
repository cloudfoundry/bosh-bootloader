package bosh

type CompilationGenerator struct{}

type Compilation struct {
	Workers             int      `yaml:"workers"`
	Network             string   `yaml:"network"`
	AZ                  string   `yaml:"az"`
	ReuseCompilationVMs bool     `yaml:"reuse_compilation_vms"`
	VMType              string   `yaml:"vm_type"`
	VMExtensions        []string `yaml:"vm_extensions"`
}

func NewCompilationGenerator() CompilationGenerator {
	return CompilationGenerator{}
}

func (CompilationGenerator) Generate() *Compilation {
	return &Compilation{
		Workers:             3,
		Network:             "private",
		AZ:                  "z1",
		ReuseCompilationVMs: true,
		VMType:              "c3.large",
		VMExtensions:        []string{"100GB_ephemeral_disk"},
	}
}
