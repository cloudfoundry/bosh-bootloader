//go:generate packr2

package cloudstack

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/gobuffalo/packr/v2"
	"strings"
)

const templatesPath = "./templates"

type TemplateGenerator struct{
	box *packr.Box
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{
		box: packr.New("cloudstack-templates", templatesPath),
	}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	vals := []string{}
	for _, name := range t.box.List() {
		val, _ := t.box.FindString(name)
		vals = append(vals, val)
	}

	return strings.Join(vals, "\n")
}
