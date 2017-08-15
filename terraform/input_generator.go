package terraform

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
	gcpInputGenerator   inputGenerator
	awsInputGenerator   inputGenerator
	azureInputGenerator inputGenerator
}

func NewInputGenerator(gcpInputGenerator inputGenerator, awsInputGenerator inputGenerator, azureInputGenerator inputGenerator) InputGenerator {
	return InputGenerator{
		gcpInputGenerator:   gcpInputGenerator,
		awsInputGenerator:   awsInputGenerator,
		azureInputGenerator: azureInputGenerator,
	}
}

func (i InputGenerator) Generate(state storage.State) (map[string]string, error) {
	switch state.IAAS {
	case "gcp":
		return i.gcpInputGenerator.Generate(state)
	case "aws":
		return i.awsInputGenerator.Generate(state)
	case "azure":
		return i.azureInputGenerator.Generate(state)
	default:
		return map[string]string{}, fmt.Errorf("invalid iaas: %q", state.IAAS)
	}
}
