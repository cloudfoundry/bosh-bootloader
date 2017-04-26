package aws

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct {
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (t TemplateGenerator) Generate(state storage.State) string {
	template := BaseTemplate

	switch state.LB.Type {
	case "concourse":
		template = strings.Join([]string{template, LBSubnetTemplate, SSLCertificateTemplate, ConcourseLBTemplate}, "\n")
	case "cf":
		if state.LB.Domain != "" {
			template = strings.Join([]string{template, LBSubnetTemplate, SSLCertificateTemplate, CFLBTemplate, CFDNSTemplate}, "\n")
		} else {
			template = strings.Join([]string{template, LBSubnetTemplate, SSLCertificateTemplate, CFLBTemplate}, "\n")
		}
	}

	return template
}
