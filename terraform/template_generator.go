package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type TemplateGenerator struct {
	gcpTemplateGenerator templateGenerator
}

func NewTemplateGenerator(gcpTemplateGenerator templateGenerator) TemplateGenerator {
	return TemplateGenerator{
		gcpTemplateGenerator: gcpTemplateGenerator,
	}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	if state.IAAS == "gcp" {
		return t.gcpTemplateGenerator.Generate(state)
	}

	return ""
}
