package aws

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type TemplateGenerator struct {
}

type SecurityGroupDescriptions struct {
	NAT               string
	Internal          string
	BOSH              string
	Concourse         string
	ConcourseInternal string
	SSHLB             string
	SSHLBInternal     string
	Router            string
	RouterInternal    string
	TCPLB             string
	TCPLBInternal     string
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (tg TemplateGenerator) Generate(state storage.State) string {
	t := BaseTemplate

	switch state.LB.Type {
	case "concourse":
		t = strings.Join([]string{t, LBSubnetTemplate, SSLCertificateTemplate, ConcourseLBTemplate}, "\n")
	case "cf":
		if state.LB.Domain != "" {
			t = strings.Join([]string{t, LBSubnetTemplate, SSLCertificateTemplate, CFLBTemplate, CFDNSTemplate}, "\n")
		} else {
			t = strings.Join([]string{t, LBSubnetTemplate, SSLCertificateTemplate, CFLBTemplate}, "\n")
		}
	}

	tmpl := template.New("descriptions")
	tmpl, err := tmpl.Parse(t)
	if err != nil {
		panic(err)
	}

	var someDesc SecurityGroupDescriptions
	if state.MigratedFromCloudFormation {
		someDesc = SecurityGroupDescriptions{
			NAT:               "NAT",
			Internal:          "Internal",
			BOSH:              "BOSH",
			Concourse:         "Concourse",
			ConcourseInternal: "Concourse Internal",
			SSHLB:             "CFSSHProxy",
			SSHLBInternal:     "CFSSHProxyInternal",
			Router:            "Router",
			RouterInternal:    "CFRouterInternal",
			TCPLB:             "CF TCP",
			TCPLBInternal:     "CF TCP Internal",
		}
	} else {
		someDesc = SecurityGroupDescriptions{
			NAT:               "NAT",
			Internal:          "Internal",
			BOSH:              "Bosh",
			Concourse:         "Concourse",
			ConcourseInternal: "Concourse Internal",
			SSHLB:             "CF SSH",
			SSHLBInternal:     "CF SSH Internal",
			Router:            "CF Router",
			RouterInternal:    "CF Router Internal",
			TCPLB:             "CF TCP",
			TCPLBInternal:     "CF TCP Internal",
		}
	}
	finalTemplate := bytes.Buffer{}

	err = tmpl.Execute(&finalTemplate, someDesc)
	if err != nil {
		panic(err)
	}

	return finalTemplate.String()
}
