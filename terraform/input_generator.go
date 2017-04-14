package terraform

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
	gcpInputGenerator inputGenerator
}

func NewInputGenerator(gcpInputGenerator inputGenerator) InputGenerator {
	return InputGenerator{
		gcpInputGenerator: gcpInputGenerator,
	}
}

func (i InputGenerator) Generate(state storage.State) (map[string]string, error) {
	if state.IAAS == "gcp" {
		return i.gcpInputGenerator.Generate(state)
	}

	return map[string]string{}, fmt.Errorf("invalid iaas: %q", state.IAAS)
}
