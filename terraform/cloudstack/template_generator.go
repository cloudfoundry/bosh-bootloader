package cloudstack

import (
	"embed"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct {
	EmbedData embed.FS
	Path      string
}

//go:embed templates
var contents embed.FS

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{
		EmbedData: contents,
		Path:      "templates",
	}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	var vals []string

	content, err := t.EmbedData.ReadDir(t.Path)
	if err != nil {
		panic(err)
	}

	for _, embedDataEntry := range content {
		vals = append(vals, embedDataEntry.Name())
	}

	return strings.Join(vals, "\n")
}
