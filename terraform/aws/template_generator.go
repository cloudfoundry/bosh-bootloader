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
		template = strings.Join([]string{template, LBSubnetTemplate, ConcourseLBTemplate}, "\n")
	case "cf":
		template = strings.Join([]string{template, LBSubnetTemplate, CFLBTemplate}, "\n")
	}

	return template
}
