package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"

type CertificateManager struct {
	CreateOrUpdateCall struct {
		CallCount int
		Receives  struct {
			Client      iam.Client
			Certificate string
			Name        string
			PrivateKey  string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *CertificateManager) CreateOrUpdate(name, certificate, privatekey string, client iam.Client) error {
	c.CreateOrUpdateCall.CallCount++
	c.CreateOrUpdateCall.Receives.Client = client
	c.CreateOrUpdateCall.Receives.Certificate = certificate
	c.CreateOrUpdateCall.Receives.PrivateKey = privatekey
	c.CreateOrUpdateCall.Receives.Name = name

	return c.CreateOrUpdateCall.Returns.Error
}
