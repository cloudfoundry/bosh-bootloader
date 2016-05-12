package iam

import (
	"io/ioutil"
	"strings"
)

type CertificateManager struct {
	certificateUploader  certificateUploader
	certificateDescriber certificateDescriber
	certificateDeleter   certificateDeleter
}

type Certificate struct {
	Name string
	Body string
	ARN  string
}

type certificateUploader interface {
	Upload(certificatePath, privateKeyPath string, iamClient Client) (string, error)
}

type certificateDescriber interface {
	Describe(certificateName string, iamClient Client) (Certificate, error)
}

type certificateDeleter interface {
	Delete(certificateName string, iamClient Client) error
}

func NewCertificateManager(certificateUploader certificateUploader, certificateDescriber certificateDescriber, certificateDeleter certificateDeleter) CertificateManager {
	return CertificateManager{
		certificateUploader:  certificateUploader,
		certificateDescriber: certificateDescriber,
		certificateDeleter:   certificateDeleter,
	}
}

func (c CertificateManager) CreateOrUpdate(name, certificatePath, privateKeyPath string, iamClient Client) (string, error) {
	if name == "" {
		return c.certificateUploader.Upload(certificatePath, privateKeyPath, iamClient)
	}

	remoteCertificate, err := c.certificateDescriber.Describe(name, iamClient)

	if err == CertificateNotFound {
		return c.certificateUploader.Upload(certificatePath, privateKeyPath, iamClient)
	}

	if err != nil {
		return "", err
	}

	localCertificateBody, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return "", err
	}

	trimmedLocalCertificateBody := strings.TrimSpace(string(localCertificateBody))

	if remoteCertificate.Body != trimmedLocalCertificateBody {
		return c.overwriteCertificate(name, certificatePath, privateKeyPath, iamClient)
	}

	return name, nil
}

func (c CertificateManager) Create(certificatePath, privateKeyPath string, iamClient Client) (string, error) {
	return c.certificateUploader.Upload(certificatePath, privateKeyPath, iamClient)
}

func (c CertificateManager) Delete(certificateName string, iamClient Client) error {
	return c.certificateDeleter.Delete(certificateName, iamClient)
}

func (c CertificateManager) Describe(certificateName string, iamClient Client) (Certificate, error) {
	return c.certificateDescriber.Describe(certificateName, iamClient)
}

func (c CertificateManager) overwriteCertificate(name, certificatePath, privateKeyPath string, iamClient Client) (string, error) {
	err := c.certificateDeleter.Delete(name, iamClient)
	if err != nil {
		return "", err
	}

	certificateName, err := c.certificateUploader.Upload(certificatePath, privateKeyPath, iamClient)
	if err != nil {
		return "", err
	}
	return certificateName, err
}
