package terraform

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OutputGenerator struct {
	gcpOutputGenerator outputGenerator
}

func NewOutputGenerator(gcpOutputGenerator outputGenerator) OutputGenerator {
	return OutputGenerator{
		gcpOutputGenerator: gcpOutputGenerator,
	}
}

func (o OutputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	if state.IAAS == "gcp" {
		return o.gcpOutputGenerator.Generate(state)
	}

	return map[string]interface{}{}, fmt.Errorf("invalid iaas: %q", state.IAAS)
}
