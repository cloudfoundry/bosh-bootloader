//go:generate packr2

package aws

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/gobuffalo/packr/v2"
)

const templatesPath = "./templates"

type templates struct {
	base           string
	iam            string
	lbSubnet       string
	cfLB           string
	cfDNS          string
	concourseLB    string
	sslCertificate string
	isoSeg         string
	vpc            string
}

type TemplateGenerator struct {
	box *packr.Box
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{
		box: packr.New("aws-templates", templatesPath),
	}
}

func (tg TemplateGenerator) Generate(state storage.State) string {
	tmpls := tg.readTemplates()
	template := strings.Join([]string{tmpls.base, tmpls.iam, tmpls.vpc}, "\n")

	switch state.LB.Type {
	case "concourse":
		template = strings.Join([]string{template, tmpls.lbSubnet, tmpls.concourseLB}, "\n")
	case "cf":
		template = strings.Join([]string{template, tmpls.lbSubnet, tmpls.cfLB, tmpls.sslCertificate, tmpls.isoSeg}, "\n")

		if state.LB.Domain != "" {
			template = strings.Join([]string{template, tmpls.cfDNS}, "\n")
		}
	}

	return template
}

func (tg TemplateGenerator) readTemplates() templates {
	listings := map[string]string{
		"base.tf":            "",
		"iam.tf":             "",
		"lb_subnet.tf":       "",
		"cf_lb.tf":           "",
		"cf_dns.tf":          "",
		"concourse_lb.tf":    "",
		"ssl_certificate.tf": "",
		"iso_segments.tf":    "",
		"vpc.tf":             "",
	}

	var errors []error
	for item := range listings {
		content, err := tg.box.FindString(item)
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
		base:           listings["base.tf"],
		iam:            listings["iam.tf"],
		lbSubnet:       listings["lb_subnet.tf"],
		cfLB:           listings["cf_lb.tf"],
		cfDNS:          listings["cf_dns.tf"],
		concourseLB:    listings["concourse_lb.tf"],
		sslCertificate: listings["ssl_certificate.tf"],
		isoSeg:         listings["iso_segments.tf"],
		vpc:            listings["vpc.tf"],
	}
}
