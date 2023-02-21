package azure

import (
	"embed"
	"fmt"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type templates struct {
	vars                 string
	resourceGroup        string
	network              string
	storage              string
	networkSecurityGroup string
	output               string
	tls                  string
	cfLB                 string
	cfDNS                string
	concourseLB          string
}

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
	tmpls := t.readTemplates()

	template := strings.Join([]string{tmpls.vars, tmpls.resourceGroup, tmpls.network, tmpls.storage, tmpls.networkSecurityGroup, tmpls.output, tmpls.tls}, "\n")

	switch state.LB.Type {
	case "cf":
		template = strings.Join([]string{template, tmpls.cfLB}, "\n")

		if state.LB.Domain != "" {
			template = strings.Join([]string{template, tmpls.cfDNS}, "\n")
		}
	case "concourse":
		template = strings.Join([]string{template, tmpls.concourseLB}, "\n")
	}

	return template
}

func (t TemplateGenerator) readTemplates() templates {
	listings := map[string]string{
		"vars.tf":                   "",
		"resource_group.tf":         "",
		"network.tf":                "",
		"storage.tf":                "",
		"network_security_group.tf": "",
		"output.tf":                 "",
		"tls.tf":                    "",
		"cf_lb.tf":                  "",
		"cf_dns.tf":                 "",
		"concourse_lb.tf":           "",
	}

	var errors []error
	for item := range listings {
		content, err := t.EmbedData.ReadDir(t.Path)
		for _, embedDataEntry := range content {
			if strings.Contains(embedDataEntry.Name(), item) {
				out, err := t.EmbedData.ReadFile(fmt.Sprintf("%s/%s", t.Path, embedDataEntry.Name()))
				if err != nil {
					errors = append(errors, err)
					break
				}
				listings[item] = string(out)
				break
			}
		}
		if err != nil {
			errors = append(errors, err)
			continue
		}
	}

	if errors != nil {
		panic(errors)
	}

	return templates{
		vars:                 listings["vars.tf"],
		resourceGroup:        listings["resource_group.tf"],
		network:              listings["network.tf"],
		storage:              listings["storage.tf"],
		networkSecurityGroup: listings["network_security_group.tf"],
		output:               listings["output.tf"],
		tls:                  listings["tls.tf"],
		cfLB:                 listings["cf_lb.tf"],
		cfDNS:                listings["cf_dns.tf"],
		concourseLB:          listings["concourse_lb.tf"],
	}
}
