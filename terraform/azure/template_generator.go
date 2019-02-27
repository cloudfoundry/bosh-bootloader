package azure

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/gobuffalo/packr/v2"
)

const templatesPath = "./templates"

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
	box *packr.Box
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{
		box: packr.New("azure-templates", templatesPath),
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
		content, err := t.box.FindString(item)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		listings[item] = content
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
