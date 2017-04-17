package terraform

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
	gcpInputGenerator inputGenerator
	awsInputGenerator inputGenerator
}

func NewInputGenerator(gcpInputGenerator inputGenerator, awsInputGenerator inputGenerator) InputGenerator {
	return InputGenerator{
		gcpInputGenerator: gcpInputGenerator,
		awsInputGenerator: awsInputGenerator,
	}
}

func (i InputGenerator) Generate(state storage.State) (map[string]string, error) {
	switch state.IAAS {
	case "gcp":
		return i.gcpInputGenerator.Generate(state)
	case "aws":
		return i.awsInputGenerator.Generate(state)
	default:
		return map[string]string{}, fmt.Errorf("invalid iaas: %q", state.IAAS)
	}
}
