package terraform

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OutputGenerator struct {
	gcpOutputGenerator outputGenerator
	awsOutputGenerator outputGenerator
}

func NewOutputGenerator(gcpOutputGenerator outputGenerator, awsOutputGenerator outputGenerator) OutputGenerator {
	return OutputGenerator{
		gcpOutputGenerator: gcpOutputGenerator,
		awsOutputGenerator: awsOutputGenerator,
	}
}

func (o OutputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	switch state.IAAS {
	case "gcp":
		return o.gcpOutputGenerator.Generate(state)
	case "aws":
		return o.awsOutputGenerator.Generate(state)
	default:
		return map[string]interface{}{}, fmt.Errorf("invalid iaas: %q", state.IAAS)
	}

}
