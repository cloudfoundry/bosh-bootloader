package vsphere

import "github.com/cloudfoundry/bosh-bootloader/storage"

type TemplateGenerator struct{}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	return ""
}
