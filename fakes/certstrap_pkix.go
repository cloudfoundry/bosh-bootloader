package fakes

import (
	"net"

	certstrappkix "github.com/square/certstrap/pkix"
)

type CertstrapPKIX struct {
	CreateCertificateAuthorityCall struct {
		CallCount int
		Receives  struct {
			Key                *certstrappkix.Key
			OrganizationalUnit string
			Years              int
			Organization       string
			Country            string
			Province           string
			Locality           string
			CommonName         string
		}
		Returns struct {
			Certificate *certstrappkix.Certificate
			Error       error
		}
	}

	CreateCertificateSigningRequestCall struct {
		CallCount int
		Receives  struct {
			Key                *certstrappkix.Key
			OrganizationalUnit string
			Years              int
			Organization       string
			Country            string
			Province           string
			Locality           string
			CommonName         string
			IpList             []net.IP
			DomainList         []string
		}
		Returns struct {
			CertificateSigningRequest *certstrappkix.CertificateSigningRequest
			Error                     error
		}
	}

	CreateCertificateHostCall struct {
		CallCount int
		Receives  struct {
			CrtAuth *certstrappkix.Certificate
			KeyAuth *certstrappkix.Key
			Csr     *certstrappkix.CertificateSigningRequest
			Years   int
		}
		Returns struct {
			Certificate *certstrappkix.Certificate
			Error       error
		}
	}
}

func (c *CertstrapPKIX) CreateCertificateAuthority(key *certstrappkix.Key, organizationalUnit string, years int, organization string, country string, province string, locality string, commonName string) (*certstrappkix.Certificate, error) {
	c.CreateCertificateAuthorityCall.CallCount++
	c.CreateCertificateAuthorityCall.Receives.Key = key
	c.CreateCertificateAuthorityCall.Receives.OrganizationalUnit = organizationalUnit
	c.CreateCertificateAuthorityCall.Receives.Years = years
	c.CreateCertificateAuthorityCall.Receives.Organization = organization
	c.CreateCertificateAuthorityCall.Receives.Country = country
	c.CreateCertificateAuthorityCall.Receives.Province = province
	c.CreateCertificateAuthorityCall.Receives.Locality = locality
	c.CreateCertificateAuthorityCall.Receives.CommonName = commonName

	return c.CreateCertificateAuthorityCall.Returns.Certificate, c.CreateCertificateAuthorityCall.Returns.Error
}

func (c *CertstrapPKIX) CreateCertificateSigningRequest(key *certstrappkix.Key, organizationalUnit string, ipList []net.IP, domainList []string, organization string, country string, province string, locality string, commonName string) (*certstrappkix.CertificateSigningRequest, error) {
	c.CreateCertificateSigningRequestCall.CallCount++

	c.CreateCertificateSigningRequestCall.Receives.Key = key
	c.CreateCertificateSigningRequestCall.Receives.OrganizationalUnit = organizationalUnit
	c.CreateCertificateSigningRequestCall.Receives.Organization = organization
	c.CreateCertificateSigningRequestCall.Receives.Country = country
	c.CreateCertificateSigningRequestCall.Receives.Province = province
	c.CreateCertificateSigningRequestCall.Receives.Locality = locality
	c.CreateCertificateSigningRequestCall.Receives.CommonName = commonName
	c.CreateCertificateSigningRequestCall.Receives.DomainList = domainList
	c.CreateCertificateSigningRequestCall.Receives.IpList = ipList

	return c.CreateCertificateSigningRequestCall.Returns.CertificateSigningRequest, c.CreateCertificateSigningRequestCall.Returns.Error
}

func (c *CertstrapPKIX) CreateCertificateHost(crtAuth *certstrappkix.Certificate, keyAuth *certstrappkix.Key, csr *certstrappkix.CertificateSigningRequest, years int) (*certstrappkix.Certificate, error) {
	c.CreateCertificateHostCall.CallCount++

	c.CreateCertificateHostCall.Receives.CrtAuth = crtAuth
	c.CreateCertificateHostCall.Receives.KeyAuth = keyAuth
	c.CreateCertificateHostCall.Receives.Csr = csr
	c.CreateCertificateHostCall.Receives.Years = years

	return c.CreateCertificateHostCall.Returns.Certificate, c.CreateCertificateHostCall.Returns.Error
}
