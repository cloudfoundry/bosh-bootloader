package aws

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var AMIs = `{
  "ap-northeast-1":  "ami-10dfc877",
  "ap-northeast-2":  "ami-1a1bc474",
  "ap-southeast-1":  "ami-36af2055",
  "ap-southeast-2":  "ami-1e91817d",
  "eu-central-1":  "ami-9ebe18f1",
  "eu-west-1":  "ami-3a849f5c",
  "eu-west-2":  "ami-21120445",
  "us-east-1":  "ami-d4c5efc2",
  "us-east-2":  "ami-f27b5a97",
  "us-gov-west-1":  "ami-c39610a2",
  "us-west-1":  "ami-b87f53d8",
  "us-west-2":  "ami-8bfce8f2"
}`

type TemplateGenerator struct{}

type TemplateData struct {
	NATDescription                 string
	InternalDescription            string
	BOSHDescription                string
	ConcourseDescription           string
	ConcourseInternalDescription   string
	SSHLBDescription               string
	SSHLBInternalDescription       string
	RouterDescription              string
	RouterInternalDescription      string
	TCPLBDescription               string
	TCPLBInternalDescription       string
	SSLCertificateNameProperty     string
	IgnoreSSLCertificateProperties string
	AWSNATAMIs                     map[string]string
}

func NewTemplateGenerator() TemplateGenerator {
	return TemplateGenerator{}
}

func (tg TemplateGenerator) Generate(state storage.State) string {
	t := BaseTemplate

	switch state.LB.Type {
	case "concourse":
		t = strings.Join([]string{t, LBSubnetTemplate, ConcourseLBTemplate, SSLCertificateTemplate}, "\n")
	case "cf":
		t = strings.Join([]string{t, LBSubnetTemplate, CFLBTemplate, SSLCertificateTemplate}, "\n")

		if state.LB.Domain != "" {
			t = strings.Join([]string{t, CFDNSTemplate, CFISOTemplate}, "\n")
		}
	}

	var ami map[string]string
	err := json.Unmarshal([]byte(AMIs), &ami)
	if err != nil {
		panic(err)
	}

	var templateData TemplateData
	if state.MigratedFromCloudFormation {
		templateData = TemplateData{
			NATDescription:               "NAT",
			InternalDescription:          "Internal",
			BOSHDescription:              "BOSH",
			ConcourseDescription:         "Concourse",
			ConcourseInternalDescription: "Concourse Internal",
			SSHLBDescription:             "CFSSHProxy",
			SSHLBInternalDescription:     "CFSSHProxyInternal",
			RouterDescription:            "Router",
			RouterInternalDescription:    "CFRouterInternal",
			TCPLBDescription:             "CF TCP",
			TCPLBInternalDescription:     "CF TCP Internal",
			SSLCertificateNameProperty:   `name              = "${var.ssl_certificate_name}"`,
			AWSNATAMIs:                   ami,
		}
	} else {
		templateData = TemplateData{
			NATDescription:               "NAT",
			InternalDescription:          "Internal",
			BOSHDescription:              "Bosh",
			ConcourseDescription:         "Concourse",
			ConcourseInternalDescription: "Concourse Internal",
			SSHLBDescription:             "CF SSH",
			SSHLBInternalDescription:     "CF SSH Internal",
			RouterDescription:            "CF Router",
			RouterInternalDescription:    "CF Router Internal",
			TCPLBDescription:             "CF TCP",
			TCPLBInternalDescription:     "CF TCP Internal",
			SSLCertificateNameProperty:   `name_prefix       = "${var.ssl_certificate_name_prefix}"`,
			AWSNATAMIs:                   ami,
		}
	}

	if state.LB.Cert == "" || state.LB.Key == "" {
		templateData.IgnoreSSLCertificateProperties = `ignore_changes = ["certificate_body", "certificate_chain", "private_key"]`
	}

	tmpl := template.New("descriptions")
	tmpl, err = tmpl.Parse(t)
	if err != nil {
		panic(err)
	}

	finalTemplate := bytes.Buffer{}

	err = tmpl.Execute(&finalTemplate, templateData)
	if err != nil {
		panic(err)
	}

	return finalTemplate.String()
}
