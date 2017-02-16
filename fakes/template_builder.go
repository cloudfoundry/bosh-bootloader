package fakes

import "github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"

type TemplateBuilder struct {
	BuildCall struct {
		Receives struct {
			KeyPairName      string
			AZs              []string
			LBType           string
			LBCertificateARN string
			IAMUserName      string
			EnvID            string
			BOSHAZ           string
		}
		Returns struct {
			Template templates.Template
		}
	}
}

func (b *TemplateBuilder) Build(keyPairName string, azs []string, lbType string, lbCertificateARN string, iamUserName string, envID string, boshAZ string) templates.Template {
	b.BuildCall.Receives.KeyPairName = keyPairName
	b.BuildCall.Receives.AZs = azs
	b.BuildCall.Receives.LBType = lbType
	b.BuildCall.Receives.LBCertificateARN = lbCertificateARN
	b.BuildCall.Receives.IAMUserName = iamUserName
	b.BuildCall.Receives.EnvID = envID
	b.BuildCall.Receives.BOSHAZ = boshAZ

	return b.BuildCall.Returns.Template
}
